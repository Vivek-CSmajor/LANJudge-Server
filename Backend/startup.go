package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// chdirToExecutableDir makes the process's own directory match the folder the
// binary actually lives in, regardless of where the OS launched it from (this
// varies by OS and by how the user double-clicked it). Every relative path
// used elsewhere in this program (server_data/..., contest-package.zip) is
// resolved against the working directory, so anchoring that directory here,
// once, at the very start of main, is what makes double-click launches behave
// the same on Windows, Mac, and Linux without touching those other paths.
func chdirToExecutableDir() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return os.Chdir(filepath.Dir(exe))
}

// ensureContestData makes sure server_data/contest/manifest.json exists
// before the rest of startup tries to read it, extracting contest-package.zip
// (expected next to the binary, i.e. in the now-current directory per
// chdirToExecutableDir) the first time the server runs. Later runs skip
// extraction entirely once the manifest is present, so restarting never
// clobbers contest data.
func ensureContestData() error {
	const manifestPath = "server_data/contest/manifest.json"
	const zipPath = "contest-package.zip"

	if _, err := os.Stat(manifestPath); err == nil {
		return nil
	}

	if _, err := os.Stat(zipPath); err != nil {
		return fmt.Errorf(
			"couldn't find contest data.\n\n" +
				"contest-package.zip must be in the same folder as this program.\n" +
				"Place it there and run this program again.",
		)
	}

	if err := os.MkdirAll("server_data/contest", 0755); err != nil {
		return fmt.Errorf("couldn't create server_data/contest: %w", err)
	}
	if err := unzip(zipPath, "server_data/contest"); err != nil {
		return fmt.Errorf("couldn't extract contest-package.zip: %w", err)
	}
	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("contest-package.zip was extracted, but manifest.json is still missing — the file may be corrupted")
	}

	log.Println("Extracted contest-package.zip into server_data/contest")
	return nil
}

// unzip extracts every entry in zipPath into destDir, rejecting any entry
// whose name would resolve outside destDir (zip-slip).
func unzip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		targetAbs, err := filepath.Abs(filepath.Join(destDir, f.Name))
		if err != nil {
			return err
		}
		if targetAbs != destAbs && !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetAbs, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
			return err
		}
		if err := extractZipFile(f, targetAbs); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(f *zip.File, targetAbs string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(targetAbs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// openDashboardWhenReady polls /health until the server is actually accepting
// connections, then opens the admin dashboard in the default browser. It
// always logs the URL up front and again once ready, so if the auto-open
// fails for any reason (unsupported OS, no default browser configured, a
// headless box), the organizer still has the URL printed in the console.
func openDashboardWhenReady(url string) {
	deadline := time.Now().Add(15 * time.Second)
	client := http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Println("Server is ready — opening the admin dashboard in your browser:", url)
				if err := openBrowser(url); err != nil {
					log.Println("Couldn't open the browser automatically (", err, ") — open this URL manually:", url)
				}
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.Println("Server didn't confirm it was ready in time — open this URL manually:", url)
}

// openBrowser launches the OS's default browser on the given URL. Each OS
// needs a different launcher: "open" on macOS, "rundll32 url.dll" on Windows
// (the standard way to invoke the URL file-association handler without a
// shell interpreting the URL's characters), and "xdg-open" on Linux desktops.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
