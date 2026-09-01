package judge
//The code in this file looks AI generated cuz its written by me but then it looked so goddamn ugly I was disgusted,bamboozled
// and impressed with my ability to write shitty code that i had to beg claude code to Reformat the whole file and make it look prettty like a princess

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RunAdHoc compiles (if needed) and runs code once against the given input,
// entirely outside the graded-submission pipeline: no DB writes, no queue,
// no verdict — just compile/run and hand back whatever happened.
func RunAdHoc(language, code, input string) (output string, errMsg string) {
	spec, ok := LanguageSpecs[language]
	if !ok {
		return "", "Unsupported language"
	}

	workspaceDir, err := os.MkdirTemp(runtimePath, "adhoc-*")
	if err != nil {
		return "", "Internal error creating workspace"
	}
	defer os.RemoveAll(workspaceDir)

	srcPath := filepath.Join(workspaceDir, spec.SourceFile)
	if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
		return "", "Internal error writing source file"
	}

	if spec.CompileCmd != "" {
		compileLog, err := runAdHocCompile(workspaceDir, spec)
		if err != nil {
			return "", compileLog
		}
	}

	inputPath := filepath.Join(workspaceDir, "input.txt")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		return "", "Internal error writing input file"
	}

	stdout, stderr, err := runDocker(0, workspaceDir, spec.RunCmd+"< /workspace/input.txt", 10*time.Second, spec)
	if err != nil {
		if stderr == "TLE" {
			return "", "Time limit exceeded"
		}
		if stderr != "" {
			return "", stderr
		}
		return "", err.Error()
	}

	return stdout, ""
}

// runAdHocCompile mirrors compile()'s docker invocation in judge.go, but
// returns the captured log directly instead of writing to a submission's
// persistent compile.log — there's no submission here to attach it to.
func runAdHocCompile(workspaceDir string, spec LanguageSpec) (string, error) {
	absWorkspaceDir, err := filepath.Abs(workspaceDir)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"docker", "run", "--rm",
		"--memory=512m",
		"--cpus=1",
		"-v", absWorkspaceDir+":/workspace",
		"-w", "/workspace",
		spec.Image,
		"bash", "-c",
		spec.CompileCmd,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "Compilation timeout", ctx.Err()
	}
	if err != nil {
		return out.String(), err
	}
	return "", nil
}
