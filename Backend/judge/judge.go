package judge

import (
	"bytes"
	"campuscode/models"
	"campuscode/utils"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type LanguageSpec struct {
	SourceFile string
	Image      string
	CompileCmd string // empty = no compile step (Python)
	RunCmd     string
}

var LanguageSpecs = map[string]LanguageSpec{
	"cpp": {
		SourceFile: "code.cpp",
		Image:      "gcc:latest",
		CompileCmd: "g++ code.cpp -O2 -std=c++17 -o program",
		RunCmd:     "./program",
	},
	"python": {
		SourceFile: "code.py",
		Image:      "python:3.12-slim",
		RunCmd:     "python3 code.py",
	},
	"java": {
		SourceFile: "Main.java",
		Image:      "eclipse-temurin:21-jdk",
		CompileCmd: "javac Main.java",
		RunCmd:     "java Main",
	},
}

func processSubmission(sub models.Submission) (string, error) {
	spec := LanguageSpecs[sub.Language]
	workspaceDir := WorkspacePathByID(sub.ID)
	err := os.MkdirAll(workspaceDir, 0755)
	if err != nil {
		return string(models.InternalError), fmt.Errorf("failed to create workspace directory: %v", err)
	}
	defer os.RemoveAll(workspaceDir)
	src := filepath.Join(SubmissionPathByID(sub.ID), spec.SourceFile)
	dst := filepath.Join(workspaceDir, spec.SourceFile)
	err = utils.CopyFile(src, dst)
	if err != nil {
		return string(models.InternalError), fmt.Errorf("failed to copy submission file: %v", err)
	}
	// compile and update db with result (if errors skip execution n display errors else continue)
	err = compile(workspaceDir, sub.ID, spec)
	if err != nil {
		log.Printf("Compilation error for submission %d is : %v", sub.ID, err)
		return string(models.CompilationError), fmt.Errorf("compilation error : %w ", err)
	}
	result, err := runAllTestCases(sub, workspaceDir, spec)
	if err != nil {
		return result, fmt.Errorf("failed to run test cases: %v", err)
	}
	return result, err
}

func runAllTestCases(sub models.Submission, workspaceDir string, spec LanguageSpec) (string, error) {
	testsDir := TestsDir(sub.ProblemID)
	testFiles, err := utils.ListInputFiles(testsDir)
	if err != nil {
		return string(models.InternalError), err
	}
	if len(testFiles) == 0 {
		return string(models.InternalError), fmt.Errorf("no test cases found for problem %d in %s", sub.ProblemID, testsDir)
	}
	for _, file := range testFiles {
		testName := filepath.Base(file)
		testNum := strings.TrimSuffix(testName, ".in")
		testInputPath := file
		testOutputPath := filepath.Join(testsDir, testNum+".out")
		result, err := runTestCase(sub.ID, testInputPath, testOutputPath, workspaceDir, spec)
		//if any test fails return WA immediately
		if err != nil {
			return result, fmt.Errorf("failed to run test case %s: %v", testNum, err)
		}
		if result == string(models.WrongAnswer) {
			//store failure.json in submission workspace with details of the failed test case for frontend to display
			return string(models.WrongAnswer), nil
		}
	}
	//if all tests pass return AC
	return string(models.Accepted), nil
}

func compile(workspaceDir string, subID int, spec LanguageSpec) error { //in docker
	submissionDir := SubmissionPathByID(subID)
	// sourcePath := filepath.Join(submissionDir,"code.cpp")
	// outputBinary := filepath.Join(submissionDir, "program")
	compileLogPath := filepath.Join(submissionDir, "compile.log")
	absWorkspaceDir, err := filepath.Abs(workspaceDir) //is unamibguous regardless of pwd, it always points to exactly one location not mattter who or where its run from.
	//its needed cuz docker daemon is a sperate process inside a vm and it doesnt share the pwd of the go server.
	if err != nil {
		return err
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
	_ = os.WriteFile(compileLogPath, out.Bytes(), 0644)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("compilation timeout")
	}
	if err != nil {
		// logBytes, _ := os.ReadFile(
		// 	filepath.Join(workspaceDir, "compile.log"),
		// )
		//return fmt.Errorf(string(logBytes))
		//implement compile.log write , write No CE if sucessfully compiles (no syntax err)
		return fmt.Errorf(out.String())
	}
	return nil
}

func runTestCase(submissionID int, testInputPath string, testExpectedOutputPath string, workspacePath string, spec LanguageSpec) (string, error) {
	inputCopyToPath := filepath.Join(workspacePath, "input.txt")
	inputBytes, err := os.ReadFile(testInputPath)
	if err != nil {
		return string(models.InternalError), err
	}
	err = os.WriteFile(inputCopyToPath, inputBytes, 0644)
	if err != nil {
		return string(models.InternalError), fmt.Errorf("failed to prepare test input: %w", err)
	}

	stdout, stderr, err := runDocker(submissionID, workspacePath, spec.RunCmd+"< /workspace/input.txt", 10*time.Second, spec)
	if err != nil {
		if stderr == "TLE" {
			return string(models.TimeLimitExceeded), nil
		}
		if strings.TrimSpace(stderr) == "" {
			return string(models.RuntimeError), fmt.Errorf("docker execution failed: %w", err)
		}
		return string(models.RuntimeError), fmt.Errorf("docker execution failed: %w; stderr: %s", err, strings.TrimSpace(stderr))
	}
	expectedBytes, err := os.ReadFile(testExpectedOutputPath)
	if err != nil {
		return string(models.InternalError), err
	}
	expected := utils.NormalizeOutput(string(expectedBytes))
	actual := utils.NormalizeOutput(stdout)
	if expected == actual {
		return string(models.Accepted), nil
	}
	return string(models.WrongAnswer), nil
}
