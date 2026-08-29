package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func ProblemsHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w,"Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	problems , err := db.GetAllProblems()
	if err != nil{
		http.Error(w,"DB error",http.StatusInternalServerError)
	}
	utils.WriteJSON(w,200,problems)
}

func ContestHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w,"Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	utils.WriteJSON(w,200,models.ActiveContest)
}

func ProblemHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w,"Invalid method called", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path,"/problems/")
	strconv.Atoi(id)

	if id == "" || id == r.URL.Path{
		http.Error(w,"id is empty in url", http.StatusBadGateway)
		return 
	}
	problemPath := filepath.Join("server_data","contest","problems",id)
	description := utils.ReadProblemStatement(problemPath)
	var problem models.Problem
	err := db.DB.First(&problem,id).Error
	if err != nil{
		http.Error(w,"internal server error 10200", http.StatusInternalServerError)
		return 
	}
	utils.WriteJSON(w,200,models.ProblemResponse{
		ID: id,
		Title: problem.Title,
		TimeLimit: problem.TimeLimit,
		MemoryLimit: problem.MemoryLimit,
		Description: description,
	})
}