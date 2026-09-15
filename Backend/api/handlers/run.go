package handlers

import (
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
)

func (h *Handlers) RunHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RunRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	output, errMsg, err := h.run.Run(req.Language, req.Code, req.Input)
	if err != nil {
		http.Error(w, "Server busy try again later babe", http.StatusTooManyRequests)
		return
	}
	utils.WriteJSON(w, 200, models.RunResponse{
		Output: output,
		Error:  errMsg,
	})
}
