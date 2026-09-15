package handlers

import (
	"campuscode/api/ws"
	"campuscode/contest"
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/service"
	"campuscode/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request body ", http.StatusBadRequest)
		return
	}
	loginCode := utils.TokenGenerator(8)
	res := models.RegisterResponse{
		Name:      req.Name,
		LoginCode: loginCode,
	}
	err = db.RegisterTeamRepository(res)
	if err != nil {
		http.Error(w, "Team may already exists", http.StatusConflict)
		return
	}
	log.Println("Team Registered")
	utils.WriteJSON(w, 200, res)
}

func (h *Handlers) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AdminLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request body bitch", http.StatusBadRequest)
		return
	}
	if h.manifest == nil ||
		req.Username != h.manifest.AdminUsername ||
		req.LoginCode != h.manifest.AdminLoginCode {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	token := utils.TokenGenerator(16)
	session := models.Sessions{
		SessionToken: token,
		Active:       true,
		Admin:        true,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(6 * time.Hour),
	}
	if err := db.CreateSession(session); err != nil {
		http.Error(w, "Internal error DB", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, map[string]string{"session_token": token})
}

func AdminGetSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	submissions, err := db.GetAllSubmissions()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, submissions)
}

func AdminGetSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	submissionID := r.PathValue("id")
	subID, err := strconv.Atoi(submissionID)
	if err != nil {
		http.Error(w, "Invalid submission id", http.StatusBadRequest)
		return
	}
	submissionPath := judge.SubmissionPathByID(subID)
	submission, err := db.GetSubmissionByID(subID)
	if err != nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}
	//submissionPath = filepath.Join(submissionPath,"/")
	spec := judge.LanguageSpecs[submission.Language]
	code, err := os.ReadFile(submissionPath + spec.SourceFile)
	if err != nil {
		http.Error(w, "Failed to read submission source", http.StatusInternalServerError)
		return
	}
	compileLog := ""
	if b, err := os.ReadFile(submissionPath + "/compile.log"); err == nil { //only populate compileLog if readfile is scuessful meaning there was something in it
		compileLog = string(b)
	}
	resp := map[string]any{
		"code":        string(code),
		"verdict":     submission.Verdict,
		"status":      submission.Status,
		"compile_log": compileLog,
		"team_id":     submission.TeamID,
		"problem_id":  submission.ProblemID,
	}
	utils.WriteJSON(w, 200, resp)
}

func (h *Handlers) AdminRejudgeSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	submissionID := r.PathValue("id")
	subID, err := strconv.Atoi(submissionID)
	if err != nil {
		http.Error(w, "Invalid submission id", http.StatusBadRequest)
		return
	}
	sub, err := db.GetSubmissionByID(subID)
	if err != nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}
	if err := db.ResetSubmissionForRejudge(subID); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	sub.Status = string(models.SubmissionQueued)
	sub.Verdict = ""

	if h.submissionQueue == nil {
		http.Error(w, "queue not initialized", http.StatusInternalServerError)
		return
	}
	h.submissionQueue <- subID

	utils.WriteJSON(w, 200, sub)
}

func AdminGetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := db.GetAllTeams()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, teams)
}

func AdminChangeTeamStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateTeamStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	if req.Status != "enabled" && req.Status != "disabled" && req.Status != "disqualified" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	teamID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}
	rows, err := db.UpdateTeamStatus(teamID, req.Status)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "team not found", http.StatusNotFound)
		return
	}
	utils.WriteJSON(w, 200, "Team Status Updated")
	//disabled teams should be not be able to view anything on their screen exept a pop up that cant be removed saying "dsiabled"
	//disqualified teams should not be able to login or submit and be immediately logged out
	//disqualified teams should be ignored in scoreboard during ordering and push them to bottom and sort on problems only on enabled teams
}

type Handlers struct {
	contestState    *contest.State
	manifest        *judge.Manifest
	submissionQueue chan int
	submissions *service.SubmissionService
}

func NewHandlers(cs *contest.State, manifest *judge.Manifest, submissionQueue chan int) *Handlers {
	return &Handlers{contestState: cs, manifest: manifest, submissionQueue: submissionQueue}
}

func (h *Handlers) AdminChangeContestState(w http.ResponseWriter, r *http.Request) {
	// should be path because its a partial update like duhhh bitch
	var req models.ChangeContestState
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	switch req.Status {
	case models.ContestPending, models.ContestActive, models.ContestPaused, models.ContestEnded:
	default:
		http.Error(w, "invalid contest status", http.StatusBadRequest)
		return
	}

	updated := h.contestState.Apply(func(c *models.Contest) {
		if req.StartTime != nil {
			c.StartTime = *req.StartTime
		}
		if req.EndTime != nil {
			c.EndTime = *req.EndTime
		}
		c.Status = req.Status
	})

	if err := db.SaveContest(&updated); err != nil { //believe me somehow its saving to the contests table
		//cuz the active contest is a pointer in memory so saving it db takes the pointer, goes to matching row with that id
		http.Error(w, "Faield to save contest", http.StatusInternalServerError)
		return
	}
	ws.BroadcastTime()
	utils.WriteJSON(w, 200, updated)
}

func AdminGetScoreboard(w http.ResponseWriter, r *http.Request) {
	ScoreboardHandler(w, r)
}

func AdminGetOverridesHandler(w http.ResponseWriter, r *http.Request) {
	overrides, err := db.GetAllOverrides()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, overrides)
}

func AdminApproveOverride(w http.ResponseWriter, r *http.Request) {
	overrideID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid override id", http.StatusBadRequest)
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	overrideRecord, err := db.GetOverrideByID(overrideID)
	if err != nil {
		http.Error(w, "Override not found", http.StatusNotFound)
		return
	}
	if overrideRecord.Status != "pending" {
		http.Error(w, "Override already resolved", http.StatusConflict)
		return
	}
	if req.Action == "approve" {
		overrideRecord.Status = "approved"
		if err := db.UpdateTeamBoundDevice(overrideRecord.TeamID, overrideRecord.NewDeviceID); err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if err := db.DeactivateTeamSessions(overrideRecord.TeamID); err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
	} else if req.Action == "reject" {
		overrideRecord.Status = "rejected"
	}
	if err := db.UpdateOverrideStatus(overrideID, overrideRecord.Status); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, overrideRecord)
}

func AdminGetContestStateHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, 200, contest.Active.Current())
}
