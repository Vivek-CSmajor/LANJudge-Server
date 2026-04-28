package judgeEngine

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

func SubmitHandler(w http.ResponseWriter, r *http.Request){
	var req SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	submissionID := saveSubmissionToDB(req) // Implement this function to save to DB and return ID
	
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
	submissionQueue <- submissionID
	if submissionQueue == nil {
		http.Error(w, "queue not initialized", 500)
    	return
	}
	writeJSON(w, 200,SubmitResponse{
		SubmissionID: submissionID,
		Status: "Queued",
	})

}


