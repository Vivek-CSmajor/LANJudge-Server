package service

import (
	"campuscode/db"
	"campuscode/models"
)

type ScoreboardService struct{}

func NewScoreboardService() *ScoreboardService {
	return &ScoreboardService{}
}

func (s *ScoreboardService) Get() ([]models.ScoreBoard, error) {
	return db.GetScoreboard()
}
