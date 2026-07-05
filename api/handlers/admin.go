package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method is called , must call POST", http.StatusMethodNotAllowed)
		return
	}
	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"Invalid Request body ", http.StatusBadRequest)
		return
	}
	loginCode := utils.TokenGenerator(8)
	res := models.RegisterResponse{
		Name : req.Name,
		LoginCode: loginCode,
	}
	err = db.RegisterTeamRepository(res);if err != nil{
		http.Error(w,"Team may already exists", http.StatusConflict) 
		return 
	}
	log.Println("Team Registered")
	utils.WriteJSON(w ,200,res)
}

func AdminGetSubmissionsHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var submissions []models.Submission
	db.DB.Find(&submissions)
	utils.WriteJSON(w,200,submissions)
}

func AdminGetSubmissionHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	submissionID := strings.TrimPrefix(r.URL.Path,"/submissions/")
	submissionPath := filepath.Join("server_data","runtime","submissions",submissionID)
	var submission models.Submission
	db.DB.Find(&submission, submissionID)
	code, err:= os.ReadFile(submissionPath+"code.cpp")
	compileLog , err := os.ReadFile(submissionPath+"compile.log")
	if err != nil{
		http.Error(w, "internal 123124", http.StatusInternalServerError)
		return
	}
	resp := map[string]any{
		"code" : string(code),
		"verdict" : submission.Verdict,
		"status" : submission.Status,
		"compile_log" : string(compileLog),
		"team_id" : submission.TeamID,
		"problem_id" : submission.ProblemID,
	}
	utils.WriteJSON(w,200,resp)
}

func AdminGetTeamsHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var teams []models.Teams
	db.DB.Find(&teams)
	utils.WriteJSON(w,200,teams)
}

func AdminChangeTeamStatusHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPatch{//cuz we are only partially updating 
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var req models.UpdateTeamStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	teamID := strings.TrimSuffix(r.URL.Path,"/teams/")
	db.DB.Model(&models.Teams{}).Where("id = ?",teamID).Update("status",req.Status)
	utils.WriteJSON(w,200,"Team Status Updated")
	//disabled teams should be not be able to view anything on their screen exept a pop up that cant be removed saying "dsiabled"
	//disqualified teams should not be able to login or submit and be immediately logged out 
	//disqualified teams should be ignored in scoreboard during ordering and push them to bottom and sort on problems only on enabled teams
}

