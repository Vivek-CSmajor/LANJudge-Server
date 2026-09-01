package handlers

import (
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
)

var semaphore = make(chan struct{},3)

func RunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var req models.RunRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var output,errMsg string
	///gotta use the semaphore here since it acts liek a buffered channel with a limit. case 1 sends empty struct (since we dont need data) 
	/// defer should be written right next to it so no matter what happens if (the logic in between if u write it) succeeds or not, defer always runs
	// suppose if the logic in bw panics then use semaphore accumulates overtime with ghost values that never free.
	select {
    case semaphore <- struct{}{}:
        defer func() { <-semaphore }()
		output, errMsg = judge.RunAdHoc(req.Language, req.Code, req.Input)
    default:
        http.Error(w, "Server busy try again later babe", http.StatusTooManyRequests)
        return
    }
	utils.WriteJSON(w, 200, models.RunResponse{
		Output: output,
		Error:  errMsg,
	})
}
