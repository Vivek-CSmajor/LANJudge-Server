package handlers

import (
	"campuscode/api/ws"
	"campuscode/contest"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/service"
	"campuscode/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request body ", http.StatusBadRequest)
		return
	}
	res, err := h.teams.Register(req)
	if err != nil {
		http.Error(w, "Team may already exists", http.StatusConflict)
		return
	}
	utils.WriteJSON(w, 200, res)
}

func (h *Handlers) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AdminLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request body bitch", http.StatusBadRequest)
		return
	}
	token, err := h.auth.AdminLogin(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal error DB", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, map[string]string{"session_token": token})
}

func (h *Handlers) AdminGetSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	submissions, err := h.submissions.ListAll()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, submissions)
}

func (h *Handlers) AdminGetSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	subID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid submission id", http.StatusBadRequest)
		return
	}
	resp, err := h.submissions.GetRaw(subID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "Submission not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to read submission source", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, resp)
}

func (h *Handlers) AdminRejudgeSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	subID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid submission id", http.StatusBadRequest)
		return
	}
	sub, err := h.submissions.Rejudge(subID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "Submission not found", http.StatusNotFound)
		default:
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, sub)
}

func (h *Handlers) AdminGetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teams.List()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, teams)
}

func (h *Handlers) AdminChangeTeamStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateTeamStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	teamID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}
	if err := h.teams.UpdateStatus(teamID, req.Status); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid status", http.StatusBadRequest)
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "team not found", http.StatusNotFound)
		default:
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, "Team Status Updated")
	//disabled teams should be not be able to view anything on their screen exept a pop up that cant be removed saying "dsiabled"
	//disqualified teams should not be able to login or submit and be immediately logged out
	//disqualified teams should be ignored in scoreboard during ordering and push them to bottom and sort on problems only on enabled teams
}

type Handlers struct {
	submissions *service.SubmissionService
	auth        *service.AuthService
	teams       *service.TeamService
	overrides   *service.OverrideService
	contest     *service.ContestService
	problems    *service.ProblemService
	scoreboard  *service.ScoreboardService
	run         *service.RunService
}

func NewHandlers(cs *contest.State, manifest *judge.Manifest, submissionQueue chan int, runSemaphore chan struct{}) *Handlers {
	return &Handlers{
		submissions: service.NewSubmissionService(submissionQueue),
		auth:        service.NewAuthService(manifest),
		teams:       service.NewTeamService(),
		overrides:   service.NewOverrideService(),
		contest:     service.NewContestService(cs),
		problems:    service.NewProblemService(),
		scoreboard:  service.NewScoreboardService(),
		run:         service.NewRunService(runSemaphore),
	}
}

func (h *Handlers) AdminChangeContestState(w http.ResponseWriter, r *http.Request) {
	// should be path because its a partial update like duhhh bitch
	var req models.ChangeContestState
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}
	updated, err := h.contest.Update(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid contest status", http.StatusBadRequest)
		default:
			http.Error(w, "Faield to save contest", http.StatusInternalServerError)
		}
		return
	}
	ws.BroadcastTime()
	utils.WriteJSON(w, 200, updated)
}

func (h *Handlers) AdminGetScoreboard(w http.ResponseWriter, r *http.Request) {
	h.ScoreboardHandler(w, r)
}

func (h *Handlers) AdminGetOverridesHandler(w http.ResponseWriter, r *http.Request) {
	overrides, err := h.overrides.List()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, overrides)
}

func (h *Handlers) AdminApproveOverride(w http.ResponseWriter, r *http.Request) {
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
	overrideRecord, err := h.overrides.Approve(overrideID, req.Action)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "Override not found", http.StatusNotFound)
		case errors.Is(err, service.ErrAlreadyResolved):
			http.Error(w, "Override already resolved", http.StatusConflict)
		default:
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, overrideRecord)
}

func (h *Handlers) AdminGetContestStateHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, 200, h.contest.Current())
}
