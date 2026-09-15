package service

import (
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrNotFound = fmt.Errorf("submission not found")
	ErrForbidden = fmt.Errorf("forbidden")
)


type SubmissionService struct {
	queue chan int 
}

func NewSubmissionService(queue chan int) *SubmissionService{
	return &SubmissionService{queue: queue}
}

const (
	StatusDisqualified = "disqualified"
	StatusDisabled     = "disabled"
)

func CanSubmit(t models.Teams) (bool, string) {
	if t.Status == nil {
		return true, ""
	}
	switch *t.Status {
	case StatusDisqualified:
		return false, "team is disqualified"
	case StatusDisabled:
		return false, "team account is disabled"
	}
	return true, ""
}

func (s *SubmissionService) Submit(req models.SubmitRequest)(int,error){
	team, err := db.GetTeamByID(req.TeamID)
	if err != nil{
		return 0, fmt.Errorf("No team with this ID error : %w",err)
	}
	if yes,reason := CanSubmit(team);!yes{
		return 0,fmt.Errorf("Team not allowed to submit reason : %w",reason)
	}
	submissionID, err := db.SaveSubmissionToDB(req) // Implement this function to save to DB and return ID
	if err != nil {
		return 0,fmt.Errorf("saving submission error : %w",err)
	}
	//store code in fs
	submissionDir := judge.SubmissionPathByID(submissionID)
	err = os.MkdirAll(submissionDir, 0755)
	if err != nil {
		return 0,fmt.Errorf("creating submission directory error : %w",err)
	}
	spec := judge.LanguageSpecs[req.Language]
	submissionPath := submissionDir + "/" + spec.SourceFile
	err = os.WriteFile(submissionPath, []byte(req.Code), 0644)
	if err != nil {
		return 0,fmt.Errorf("writing submission code error : %w",err)
	}
	// push to queue
	if s.queue == nil { //exists but aint containing nuthin
		return 0, fmt.Errorf("queue not initiaized")
	}
	s.queue <- submissionID
	return submissionID, nil
}

func (s *SubmissionService) GetDetail(submissionID string, teamID int)(models.SubmissionDetail,error){
	submission, err := db.GetSubmissionByRawID(submissionID)
	if err != nil {
		return models.SubmissionDetail{},ErrNotFound
	}
	if submission.TeamID != teamID {
		return models.SubmissionDetail{}, ErrForbidden
	}
	code := ""
	if spec, ok := judge.LanguageSpecs[submission.Language]; ok {
		codePath := filepath.Join(judge.SubmissionPathByID(submission.ID), spec.SourceFile)
		if b, err := os.ReadFile(codePath); err == nil {
			code = string(b)
		}
	}
	return models.SubmissionDetail{
		Submission: submission,
		Code: code,
	},nil
}