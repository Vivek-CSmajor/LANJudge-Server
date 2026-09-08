package judge

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"

	"time"
)

func runDocker(submissionID int, workspacePath string, command string, timeoout time.Duration, spec LanguageSpec) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoout) //here timeout is passed into it from the problems Time limit
	defer cancel()
	absWorkspacePath, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", "", err
	}
	cmd := exec.CommandContext(
		ctx,
		"docker", "run", "--rm",
		"--memory=256m",
		"--cpus=0.5",
		"--network=none",
		"--pids-limit=64",
		"-v", absWorkspacePath+":/workspace",
		"-w", "/workspace",
		spec.Image,
		"bash", "-c", command,
	)
	var stdout, stderr bytes.Buffer //two in memory buffers , growable bytes arrays
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout //redirecting the standard input output streams into the self made buffer
	// stdout.String() // contains normal output
	// stderr.String() // contains errors/logs

	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "", "TLE", ctx.Err()
	}
	if err != nil {
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}
