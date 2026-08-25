package handlers

import (
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
	"os"
)

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { //enforces post only for submission
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	team := TeamFromContext(r)
	req.TeamID = team.ID
	submissionID := db.SaveSubmissionToDB(req) // Implement this function to save to DB and return ID
	//store code in fs
	submissionDir := judge.SubmissionPathByID(submissionID)
	err = os.MkdirAll(submissionDir, 0755)
	if err != nil {
		http.Error(w, "Failed to create submission directory", http.StatusInternalServerError)
		return
	}
	submissionPath := submissionDir + "/code.cpp"
	err = os.WriteFile(submissionPath, []byte(req.Code), 0644)
	if err != nil {
		http.Error(w, "Failed to save submission code", http.StatusInternalServerError)
		return
	}
	// push to queue
	if judge.SubmissionQueue == nil {//exists but aint containing nuthin
		http.Error(w, "queue not initialized", 500)
		return
	}
	judge.SubmissionQueue <- submissionID
	
	utils.WriteJSON(w, 200, models.SubmitResponse{
		SubmissionID: submissionID,
		Status:       "Queued",
	})
}

func GetSubmissionHandler(w http.ResponseWriter,r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	//id := strings.TrimPrefix(r.URL.Path,"/submissions/")
	id := r.PathValue("id")
	var submission models.Submission
	err := db.DB.First(&submission,id).Error
	if err != nil{
		http.Error(w,"Submission Not found",http.StatusNotFound)
		return
	}
	team := TeamFromContext(r)
	if submission.TeamID != team.ID {
		http.Error(w,"Forbidden",http.StatusForbidden)
		return
	}
	utils.WriteJSON(w,200,submission)
}
