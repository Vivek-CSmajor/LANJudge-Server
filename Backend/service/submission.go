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
	ErrNotFound  = fmt.Errorf("submission not found")
	ErrForbidden = fmt.Errorf("forbidden")
)

type SubmissionService struct {
	queue chan int
}

func NewSubmissionService(queue chan int) *SubmissionService {
	return &SubmissionService{queue: queue}
}

func (s *SubmissionService) Submit(req models.SubmitRequest) (int, error) {
	team, err := db.GetTeamByID(req.TeamID)
	if err != nil {
		return 0, fmt.Errorf("No team with this ID error : %w", err)
	}
	if yes, reason := team.CanSubmit(); !yes {
		return 0, fmt.Errorf("Team not allowed to submit reason : %w", reason)
	}
	submissionID, err := db.SaveSubmissionToDB(req) // Implement this function to save to DB and return ID
	if err != nil {
		return 0, fmt.Errorf("saving submission error : %w", err)
	}
	//store code in fs
	submissionDir := judge.SubmissionPathByID(submissionID)
	err = os.MkdirAll(submissionDir, 0755)
	if err != nil {
		return 0, fmt.Errorf("creating submission directory error : %w", err)
	}
	spec := judge.LanguageSpecs[req.Language]
	submissionPath := submissionDir + "/" + spec.SourceFile
	err = os.WriteFile(submissionPath, []byte(req.Code), 0644)
	if err != nil {
		return 0, fmt.Errorf("writing submission code error : %w", err)
	}
	// push to queue
	if s.queue == nil { //exists but aint containing nuthin
		return 0, fmt.Errorf("queue not initiaized")
	}
	s.queue <- submissionID
	return submissionID, nil
}

func (s *SubmissionService) GetDetail(submissionID string, teamID int) (models.SubmissionDetail, error) {
	submission, err := db.GetSubmissionByRawID(submissionID)
	if err != nil {
		return models.SubmissionDetail{}, ErrNotFound
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
		Code:       code,
	}, nil
}

func (s *SubmissionService) ListForTeam(teamID int, problemID string) ([]models.Submission, error) {
	return db.GetSubmissionsForTeam(teamID, problemID)
}

func (s *SubmissionService) ListAll() ([]models.Submission, error) {
	return db.GetAllSubmissions()
}

func (s *SubmissionService) GetRaw(id int) (models.AdminSubmissionDetail, error) {
	submission, err := db.GetSubmissionByID(id)
	if err != nil {
		return models.AdminSubmissionDetail{}, ErrNotFound
	}
	submissionPath := judge.SubmissionPathByID(id)
	spec := judge.LanguageSpecs[submission.Language]
	code, err := os.ReadFile(submissionPath + spec.SourceFile)
	if err != nil {
		return models.AdminSubmissionDetail{}, fmt.Errorf("failed to read submission source : %w", err)
	}
	compileLog := ""
	if b, err := os.ReadFile(submissionPath + "/compile.log"); err == nil {
		compileLog = string(b)
	}
	return models.AdminSubmissionDetail{
		Code:       string(code),
		Verdict:    submission.Verdict,
		Status:     submission.Status,
		CompileLog: compileLog,
		TeamID:     submission.TeamID,
		ProblemID:  submission.ProblemID,
	}, nil
}

func (s *SubmissionService) Rejudge(id int) (models.Submission, error) {
	sub, err := db.GetSubmissionByID(id)
	if err != nil {
		return models.Submission{}, ErrNotFound
	}
	if err := db.ResetSubmissionForRejudge(id); err != nil {
		return models.Submission{}, err
	}
	sub.Status = string(models.SubmissionQueued)
	sub.Verdict = ""
	if s.queue == nil {
		return models.Submission{}, fmt.Errorf("queue not initialized")
	}
	s.queue <- id
	return sub, nil
}
