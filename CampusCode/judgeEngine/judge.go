package judgeEngine

import (
	//"database/sql"
	//"fmt"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	//"gorm.io/gorm"
	//"github.com/gin-gonic/gin"
)

// This File contains logic for judging one submission

// Edit sends when submit button is clicked in the frontend
// Backend receives metadata + code in JSON format --> takes the code away and stores it in the file system --> compiles it --> runs it against test cases --> returns results to the frontend
// if output for any test case != expected output --> return "wrong asnwer" else return "accepted"

// This is the Main Function that does everything by calling other funcs
func JudgeSubmission(sub Submission, expectedOutput string)(string,error){
	// Store the code in the file system
	filepath, err := storeSubmission(sub.Code, sub.ID)
	if err != nil {
		return "", err
	}

	// Run and judge
	result, err := runAndJudge(filepath, expectedOutput)
	if err != nil {
		return result, err
	}
	
	log.Println(result)
	return "Judging completed"+result, nil
}



func storeSubmission(code string, id int) (string, error){
	err := os.MkdirAll("submissions", 0755) 
	if err != nil {
		return "", err
	}
	fileName := filepath.Join("submissions",strconv.Itoa(id)+".cpp")
	err2 := os.WriteFile(fileName,[]byte(code), 0644)
	if err2 != nil {
		return "", err
	}
	return fileName, err

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

func runAndJudge(fileName string,expectedOutput string) (result string, err error){
	exe, err := compile(fileName)
	if err != nil {
		return "compilation err", err
	}
	//run but provide input.txt 
	inputFile, err := os.Open("input.txt")
	if err != nil {
		return "", fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()
	cmd := exec.Command(exe)
	cmd.Stdin = inputFile
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Runtime err", err
	}
	// Do something with the compiled output
	// expectedOutput, _ := os.ReadFile("expected_output.txt")
	if strings.TrimSpace(string(output)) == strings.TrimSpace(string(expectedOutput)){
		result = "Accepted :)"
		return result, nil
	}
	result = "Wrong Answer :("
	return result, nil
}