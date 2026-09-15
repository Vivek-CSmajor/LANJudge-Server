package service

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"log"
	"path/filepath"
)

type ProblemService struct{}

func NewProblemService() *ProblemService {
	return &ProblemService{}
}

func (s *ProblemService) List() ([]models.Problem, error) {
	return db.GetAllProblems()
}

func (s *ProblemService) GetByID(id string) (models.ProblemResponse, error) {
	problemPath := filepath.Join("server_data", "contest", "problems", id)
	description, err := utils.ReadProblemStatement(problemPath)
	if err != nil {
		log.Printf("failed to read statement.md for problem %s: %v", id, err)
	}
	problem, err := db.GetProblemByID(id)
	if err != nil {
		return models.ProblemResponse{}, ErrNotFound
	}
	return models.ProblemResponse{
		ID:          id,
		Title:       problem.Title,
		TimeLimit:   problem.TimeLimit,
		MemoryLimit: problem.MemoryLimit,
		Description: description,
	}, nil
}
