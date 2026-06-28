package models

import (
	"time"
)

type Submission struct{
	ID int `json:"id" gorm:"primaryKey"`
	TeamID int `json:"team_id"`
	ProblemID int `json:"problem_id"`
	//Code string `json:"code"`
	Status string `json:"status"`
	Verdict SubmissionVerdict `json:"verdict"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type Teams struct{
	ID int `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique"`
	//Members string `json:"members"` 
	LoginCode string `json:"login_code" gorm:"unique"`
	//CreatedAt time.Time `json:"created_at"`
	BoundDeviceID *string `json:"bound_device_id"`
	Status *string `json:"status"` // can be disabled,disqualified,enabled
}

type Sessions struct{
	//ID int `json:"id" gorm:"primaryKey"`
	TeamID int `json:"team_id" gorm:"not null;index"`
	DeviceID string `json:"device_id"`
	SessionToken string `json:"session_token" gorm:"primaryKey"`//new shyte: primary keys is. already unique , like obviously i didnt think of that cuz i m way too dumb 
	Active bool `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

//why tf am i still doing this project
type DeviceOverride struct{
	ID int `json:"id" gorm:"primaryKey"`
	TeamID int `json:"team_id" gorm:"not null;index"`
	NewDeviceID string `json:"new_device_id"`
	Status string `json:"status"` // "pending", "approved", "rejected"
	CreatedAt time.Time `json:"created_at"`
}



type Contest struct{
	ID int `json:"id"`
	Title string `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime time.Time `json:"end_time"`
	Status string `json:"status"`
}

type TeamProblemStats struct {
	//ID int `gorm:"primaryKey"`
	// TeamID int `json:"team_id" gorm:"uniqueIndex:idx_team_problem"`//composite uniqueness on team and problem.
    //ProblemID int `json:"problem_id" gorm:"uniqueIndex:idx_team_problem"`
	TeamID int `gorm:"primaryKey"`
	ProblemID int `gorm:"primaryKey"`
	WrongAttemptCount int `json:"wrong_attempt_count"`
	Solved bool `json:"solved"`
	FirstAcTime *time.Time `json:"first_ac_time"`
	Penalty *int `json:"penalty"`
}

type ScoreBoard struct {
	TeamID int `json:"team_id" gorm:"primaryKey"`
	TotalSolved int `json:"total_solved"`
	TotalPenalty int `json:"total_penalty"`
	LastAcTime time.Time `json:"last_ac_time"`
	Rank int `json:"rank"`
}

type Problem struct {
	ID int `json:"id" gorm:"primaryKey"`
	Title string `json:"title"`
	TimeLimit time.Duration `json:"time_limit"`
	MemoryLimit int `json:"memory_limit"`
	SolvedCount int `json:"solved_count"`
}

type ProblemResponse struct {
    ID          string        `json:"id"`
    Title       string        `json:"title"`
    TimeLimit   time.Duration `json:"time_limit"`
    MemoryLimit int           `json:"memory_limit"`
    Description string        `json:"description"`
}

