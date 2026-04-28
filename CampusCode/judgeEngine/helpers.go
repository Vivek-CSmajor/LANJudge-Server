package judgeEngine

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// func getSubmissionPath(subID int) string{
// 	return filepath.Join("submissions", strconv.Itoa(subID)+".cpp")
// }

func saveSubmissionToDB(sub SubmitRequest) int{ // saves submission data to db and return the subID
	submission := Submission{
		UserID: sub.UserID,
		ProblemID: sub.ProblemID,
		Code: sub.Code,
		Status : "Queued",	
	}
	DB.Create(&submission)
	return submission.ID
}

func getSubmissionByID(subID int) (Submission, error){
	var sub Submission 
	err := DB.First(&sub, subID).Error
	if err != nil {
		return Submission{},err
	}
	return sub, nil

}

func writeJSON(w http.ResponseWriter, status int, a any){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(a)
}

func copyFile(src string, dst string) error{
	input , err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()
	output , err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()
	_, _ = io.Copy(output,input)
	return err
}

func listInputFiles(dir string) []string{
	var files []string
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries{ //range returns index and value
		name := entry.Name()
		if strings.HasSuffix(name, ".in") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files
}