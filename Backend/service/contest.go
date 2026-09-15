package service

import (
	"campuscode/contest"
	"campuscode/db"
	"campuscode/models"
)

type ContestService struct {
	state *contest.State
}

func NewContestService(state *contest.State) *ContestService {
	return &ContestService{state: state}
}

func (s *ContestService) Current() models.Contest {
	return s.state.Current()
}

func (s *ContestService) Update(req models.ChangeContestState) (models.Contest, error) {
	switch req.Status {
	case models.ContestPending, models.ContestActive, models.ContestPaused, models.ContestEnded:
	default:
		return models.Contest{}, ErrInvalidInput
	}
	updated := s.state.Apply(func(c *models.Contest) {
		if req.StartTime != nil {
			c.StartTime = *req.StartTime
		}
		if req.EndTime != nil {
			c.EndTime = *req.EndTime
		}
		c.Status = req.Status
	})
	if err := db.SaveContest(&updated); err != nil {
		return models.Contest{}, err
	}
	return updated, nil
}
