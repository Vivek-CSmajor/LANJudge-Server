package models

import "time"

type SubmitRequest  struct { 
	TeamID int `json:"team_id"`
	ProblemID int `json:"problem_id"`
	Code string `json:"code"`
	//Input string `json:"input"`
}

type SubmitResponse struct {
	SubmissionID int `json:"submission_id"`
	Status string `json:"status"`
}

type RegisterRequest struct {
	Name string `json:"name"`
	Members string `json:"members"` 
}

type RegisterResponse struct {
	Name string `json:"name"`
	LoginCode string `json:"login_code"`
}

type LoginRequest struct {
	Name string `json:"name"`
	LoginCode string `json:"login_code"`
	BoundDeviceID string `json:"bound_device_id"`
}

type LoginResponse struct {
	SessionToken string `json:"session_token"`	
}

type UpdateTeamStatusRequest struct {
	Status string `json:"status"`
}

type ChangeContestState struct {
	StartTime time.Time `json:"start_time"`
	EndTime time.Time	`json:"end_time"`
	Status ContestStatus `json:"status"`
}

