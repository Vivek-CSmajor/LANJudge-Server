package handlers

import (
	"campuscode/db"
	"campuscode/judgeEngine"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

func SubmitHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {//enforces post only for submission
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	submissionID := db.SaveSubmissionToDB(req) // Implement this function to save to DB and return ID
	
	//store code in fs
	submissionDir := "server_data/runtime/submissions/"+strconv.Itoa(submissionID)
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
	judgeEngine.SubmissionQueue <- submissionID
	if judgeEngine.SubmissionQueue == nil {
		http.Error(w, "queue not initialized", 500)
    	return
	}
	utils.WriteJSON(w, 200,models.SubmitResponse{
		SubmissionID: submissionID,
		Status: "Queued",
	})
}
