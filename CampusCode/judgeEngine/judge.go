package judgeEngine

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
	"strconv"
	"strings"
	"time"
)

func processSubmission(sub models.Submission)(string,error){
	workspaceDir := filepath.Join("server_data","runtime","workspaces",strconv.Itoa(sub.ID))
	err := os.MkdirAll(workspaceDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %v", err)
	}
	defer os.RemoveAll(workspaceDir)
	src := filepath.Join("server_data","runtime","submissions",strconv.Itoa(sub.ID),"code.cpp")
	dst := filepath.Join(workspaceDir,"code.cpp")
	err = utils.CopyFile(src,dst)
	if err != nil {
		return "", fmt.Errorf("failed to copy submission file: %v", err)
	}
	// compile and update db with result (if errors skip execution n display errors else continue)
	err = compile(workspaceDir,sub.ID)
	if err != nil {
		log.Printf("Compilation error for submission %d is : %v",sub.ID,err)
		return string(models.CompilationError), fmt.Errorf("compilation error : %w ", err)
	}
	result, err := runAllTestCases(sub, workspaceDir)
	if err != nil {
		return "", fmt.Errorf("failed to run test cases: %v", err)
	}
	return result,err 
}

func runAllTestCases(sub models.Submission,workspaceDir string) (string,error){
	testsDir := filepath.Join("server_data","contest","problems",strconv.Itoa(sub.ProblemID),"tests")
	testFiles, err := utils.ListInputFiles(testsDir)
	if err != nil{
		return "RE",err
	}
	log.Println("TEST DIR:", testsDir)
	log.Println("TEST FILES:", testFiles)
	for _, file := range testFiles{
		//testNum := strings.TrimSuffix(file, ".in")
		testName := filepath.Base(file)
		testNum := strings.TrimSuffix(testName,".in")
		testInputPath := file
		testOutputPath := filepath.Join(testsDir, testNum+".out")
		result, err := runTestCase(sub.ID,testInputPath,testOutputPath,workspaceDir)
		//if any test fails return WA immediately 
		if err != nil {
			return string(models.RuntimeError), fmt.Errorf("failed to run test case %s: %v", testNum, err)
		}
		if result == string(models.WrongAnswer) {
			//store failure.json in submission workspace with details of the failed test case for frontend to display
			return string(models.WrongAnswer) , nil
		}
	}
	//if all tests pass return AC
	return string(models.Accepted) , nil
}


func compile(workspaceDir string, subID int)error{ //in docker
submissionDir := filepath.Join("server_data","runtime","submissions",strconv.Itoa(subID))
	// sourcePath := filepath.Join(submissionDir,"code.cpp")
	// outputBinary := filepath.Join(submissionDir, "program")
	compileLogPath := filepath.Join(submissionDir, "compile.log")
	absWorkspaceDir , err := filepath.Abs(workspaceDir)
	if err != nil {
		return err
	}
	ctx , cancel := context.WithTimeout( context.Background(), 30*time.Second )
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		"docker", "run", "--rm",
		"--memory=512m",
		"--cpus=1",
		"-v", absWorkspaceDir+":/workspace",
		"-w", "/workspace",
		"gcc:latest",
		"bash", "-c",
		"g++ code.cpp -O2 -std=c++17 -o program",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	_ = os.WriteFile(compileLogPath,out.Bytes(),0644)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("compilation timeout")
	}
	if err != nil {
		logBytes, _ := os.ReadFile(
			filepath.Join(workspaceDir, "compile.log"),
		)
		return fmt.Errorf(string(logBytes))
	}
	return nil
}




func runTestCase(submissionID int, testInputPath string, testExpectedOutputPath string, workspacePath string) (string, error){
	inputCopyToPath := filepath.Join(workspacePath,"input.txt")
	inputBytes, err := os.ReadFile(testInputPath)
	if err != nil {
    return "", err
	}
	err = os.WriteFile(inputCopyToPath,inputBytes,0644)
	if err != nil {
		return "", fmt.Errorf("failed to prepare test input: %w", err)
	}

	stdout,stderr,err := runDocker(submissionID, workspacePath, "./program < /workspace/input.txt",2*time.Second)
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
		return "exceptedBytes error in runTestCase", err
	}
	expected := utils.NormalizeOutput(string(expectedBytes))
	actual := utils.NormalizeOutput(stdout)
	if expected == actual {
		return string(models.Accepted), nil
	}
	return string(models.WrongAnswer), nil
}

