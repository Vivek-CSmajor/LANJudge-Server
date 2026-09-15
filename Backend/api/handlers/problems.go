package handlers

import (
	"campuscode/utils"
	"net/http"
	"strconv"
)

func (h *Handlers) ProblemsHandler(w http.ResponseWriter, r *http.Request) {
	problems, err := h.problems.List()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
	}
	utils.WriteJSON(w, 200, problems)
}

func (h *Handlers) ContestHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, 200, h.contest.Current())
}

func (h *Handlers) ProblemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	strconv.Atoi(id)

	if id == "" {
		http.Error(w, "id is empty in url", http.StatusBadRequest)
		return
	}
	resp, err := h.problems.GetByID(id)
	if err != nil {
		http.Error(w, "internal server error 10200", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, resp)
}
