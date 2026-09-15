package service

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"log"
)

type TeamService struct{}

func NewTeamService() *TeamService {
	return &TeamService{}
}

func (s *TeamService) Register(req models.RegisterRequest) (models.RegisterResponse, error) {
	loginCode := utils.TokenGenerator(8)
	res := models.RegisterResponse{
		Name:      req.Name,
		LoginCode: loginCode,
	}
	if err := db.RegisterTeamRepository(res); err != nil {
		return models.RegisterResponse{}, ErrConflict
	}
	log.Println("Team Registered")
	return res, nil
}

func (s *TeamService) List() ([]models.Teams, error) {
	return db.GetAllTeams()
}

func (s *TeamService) UpdateStatus(teamID int, status string) error {
	switch status {
	case models.TeamEnabled, models.TeamDisabled, models.TeamDisqualified:
	default:
		return ErrInvalidInput
	}
	rows, err := db.UpdateTeamStatus(teamID, status)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
