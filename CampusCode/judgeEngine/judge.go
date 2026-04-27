package judgeEngine

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"strings"
)

func processSubmission(sub Submission)(string,error){
	workspaceDir := filepath.Join("server_data","runtime","workspaces",strconv.Itoa(sub.ID))
	err := os.MkdirAll(workspaceDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %v", err)
	}

	src := filepath.Join("server_data","runtime","submissions",strconv.Itoa(sub.ID),"code.cpp")
	dst := filepath.Join(workspaceDir,"code.cpp")
	err = copyFile(src,dst)
	if err != nil {
		return "", fmt.Errorf("failed to copy submission file: %v", err)
	}
	// compile and update db with result (if errors skip execution n display errors else continue)
	exe , err := compile(dst)
	if err != nil {
		log.Printf("Compilation error for submission %d is : %v",sub.ID,err)
		return "", fmt.Errorf("compilation error : %w ", err)
	}
	//output becomes : server_data/runtime/workspaces/<id>/code.cpp.out
	
	result, err := runAllTestCases(sub, workspaceDir, exe)
	if err != nil {
		return "", fmt.Errorf("failed to run test cases: %v", err)
	}

	os.RemoveAll(workspaceDir)
	return result,err 
}

func runAllTestCases(sub Submission,workspaceDir,execPath string) (string,error){
	testsDir := filepath.Join("server_data","contest","problems",strconv.Itoa(sub.ProblemID),"tests")
	testFiles := listInputFiles(testsDir)
	for _, file := range testFiles{
		testNum := strings.TrimSuffix(file, ".in")

		testInputPath := filepath.Join(testsDir, file)
		testOutputPath := filepath.Join(testsDir, testNum+".out")
		result, err := runTestCase(testInputPath,testOutputPath,workspaceDir,execPath)
		if err != nil {
			return "", fmt.Errorf("failed to run test case %s: %v", testNum, err)
		}
		if result == "Correct" {
			return "Correct", nil
		}
		if result == "Wrong answer" {
			//store failure.json in submission workspace with details of the failed test case for frontend to display
			return result , nil
		}
	}
	return "Accepted", nil
}

func compile(fileName string)(string,error){
	outputExe := fileName + ".out"
	// Compile the code using g++
	cmd := exec.Command("g++",fileName,"-Wall","-o",outputExe)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compile err : %s",string(out))
	}
	return outputExe, nil
}

func runTestCase(testInputPath,testOutputPath,workspacePath,execPath string) (string,error){
	outputPath := filepath.Join(workspacePath, "output.txt")
	inputFile, err := os.Open(testInputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	cmd := exec.Command(execPath)
	cmd.Dir = workspacePath
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Runtime error: %v", err)
	}
	//comparing outputs
	expectedBytes , _ := os.ReadFile(testOutputPath)
	actualBytes , _ := os.ReadFile(outputPath)

	expected := normalizeOutput(string(expectedBytes))
	actual := normalizeOutput(string(actualBytes))
	
	if expected == actual {
		return "Correct", nil
	}
	return "Wrong answer", nil
}

func normalizeOutput(s string) string{
	s = strings.ReplaceAll(s,"\r\n","\n")//convert windwos line endings to unix
	lines := strings.Split(s,"\n")
	for i:= range lines{
		lines[i] = strings.TrimRight(lines[i],"\t") // remove trailing tabs
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.TrimSpace(strings.Join(lines,"\n"))
}