package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func ProblemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	problems, err := db.GetAllProblems()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
	}
	utils.WriteJSON(w, 200, problems)
}

func ContestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	utils.WriteJSON(w, 200, models.ActiveContest)
}

func ProblemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	strconv.Atoi(id)

	if id == "" {
		http.Error(w, "id is empty in url", http.StatusBadRequest)
		return
	}
	problemPath := filepath.Join("server_data", "contest", "problems", id)
	description, err := utils.ReadProblemStatement(problemPath)
	if err != nil {
		log.Printf("failed to read statement.md for problem %s: %v", id, err)
	}
	problem, err := db.GetProblemByID(id)
	if err != nil {
		http.Error(w, "internal server error 10200", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, models.ProblemResponse{
		ID:          id,
		Title:       problem.Title,
		TimeLimit:   problem.TimeLimit,
		MemoryLimit: problem.MemoryLimit,
		Description: description,
	})
}
