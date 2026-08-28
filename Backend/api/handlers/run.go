package handlers

import (
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
)

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
	output, errMsg := judge.RunAdHoc(req.Language, req.Code, req.Input)
	utils.WriteJSON(w, 200, models.RunResponse{
		Output: output,
		Error:  errMsg,
	})
}
