package handlers

import (
	"campuscode/api/ws"
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
	submissionID := r.PathValue("id")
	subID,_ := strconv.Atoi(submissionID)
	submissionPath := judge.SubmissionPathByID(subID)
	var submission models.Submission
	db.DB.Find(&submission, submissionID)
	//submissionPath = filepath.Join(submissionPath,"/")
	spec := judge.LanguageSpecs[submission.Language]
	code, err:= os.ReadFile(submissionPath+spec.SourceFile)
	compileLog , err := os.ReadFile(submissionPath+"/compile.log")
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

func AdminRejudgeSubmissionHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	submissionID := r.PathValue("id")
	subID,_ := strconv.Atoi(submissionID)
	sub, err := db.GetSubmissionByID(subID)
	if err != nil{
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}
	sub.Status = "Queued"
	sub.Verdict = ""
	db.DB.Save(&sub)

	if judge.SubmissionQueue == nil{
		http.Error(w, "queue not initialized", http.StatusInternalServerError)
		return
	}
	judge.SubmissionQueue <- subID

	utils.WriteJSON(w,200,sub)
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
	teamID := r.PathValue("id")
	db.DB.Model(&models.Teams{}).Where("id = ?",teamID).Update("status",req.Status)
	utils.WriteJSON(w,200,"Team Status Updated")
	//disabled teams should be not be able to view anything on their screen exept a pop up that cant be removed saying "dsiabled"
	//disqualified teams should not be able to login or submit and be immediately logged out 
	//disqualified teams should be ignored in scoreboard during ordering and push them to bottom and sort on problems only on enabled teams
}

func AdminChangeContestState(w http.ResponseWriter, r *http.Request){
	// should be path because its a partial update like duhhh bitch
	if r.Method != http.MethodPatch{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var req models.ChangeContestState
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}	
	models.ActiveContest.StartTime = req.StartTime
	models.ActiveContest.EndTime = req.EndTime
	models.ActiveContest.Status = req.Status
	db.DB.Save(models.ActiveContest)//believe me somehow its saving to the contests table
	//cuz the active contest is a pointer in memory so saving it db takes the pointer, goes to matching row with that id 
	ws.BroadcastTime()
	utils.WriteJSON(w,200,models.ActiveContest)
}

func AdminGetScoreboard(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	ScoreboardHandler(w,r)
}

func AdminGetOverridesHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	var overrides []models.DeviceOverride
	db.DB.Find(&overrides)
	utils.WriteJSON(w,200,overrides)
}

func AdminApproveOverride(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	var req struct {
		Action string `json:"action"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	var overrideRecord models.DeviceOverride
	db.DB.First(&overrideRecord,"id = ?",id)
	if req.Action == "approve"{
		overrideRecord.Status = "approved"
		db.DB.Model(&models.Teams{}).Where("id = ?",overrideRecord.TeamID).Update("bound_device_id",overrideRecord.NewDeviceID)
		db.DB.Model(&models.Sessions{}).Where("team_id = ?",overrideRecord.TeamID).Update("active",false)
	}else if req.Action == "reject"{
		overrideRecord.Status = "rejected"
	}
	db.DB.Save(&overrideRecord)
	utils.WriteJSON(w,200,overrideRecord)
}