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
	token := r.Header.Get("X-Session-Token")
	var session models.Sessions
	db.DB.Where("session_token = ? AND active = true",token).First(&session)
	req.TeamID = session.TeamID
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
	}
	utils.WriteJSON(w,200,submission)
}
