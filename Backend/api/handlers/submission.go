package handlers

import (
	"campuscode/models"
	"campuscode/service"
	"campuscode/utils"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *Handlers) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	team := TeamFromContext(r)
	req.TeamID = team.ID
	submissionID, err := h.submissions.Submit(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, models.SubmitResponse{
		SubmissionID: submissionID,
		Status:       "Queued",
	})
}

func (h *Handlers) GetSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	team := TeamFromContext(r)
	detail, err := h.submissions.GetDetail(id, team.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "Submission Not Found", http.StatusNotFound)
		case errors.Is(err, service.ErrForbidden):
			http.Error(w, "Forbidden", http.StatusForbidden)
		default:
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, detail)
}

func (h *Handlers) ListMySubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	team := TeamFromContext(r)

	submissions, err := h.submissions.ListForTeam(team.ID, r.URL.Query().Get("problem_id"))
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, submissions)
}
