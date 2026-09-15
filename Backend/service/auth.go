package service

import (
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"fmt"
	"log"
	"time"
)

type AuthService struct {
	manifest *judge.Manifest
}

func NewAuthService(manifest *judge.Manifest) *AuthService {
	return &AuthService{manifest: manifest}
}

func (s *AuthService) Login(req models.LoginRequest, deviceID string) (string, error) {
	team, err := db.GetTeamByNameAndLoginCode(req.Name, req.LoginCode)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if team.Status != nil && *team.Status == StatusDisqualified {
		return "", ErrDisqualified
	}
	//3 cases
	//Case 1 : First time logging in so no existing session
	if team.BoundDeviceID == nil {
		log.Println("Babes first time log in ")
		if err := db.UpdateTeamBoundDevice(team.ID, deviceID); err != nil {
			return "", fmt.Errorf("failed to bind device : %w", err)
		}
		return s.startSession(team.ID, deviceID)
	}
	//Case2 : Login again but with diff device
	if *team.BoundDeviceID != deviceID {
		//call override
		if err := db.CreateDeviceOverride(models.DeviceOverride{TeamID: team.ID, NewDeviceID: deviceID, Status: "pending", CreatedAt: time.Now()}); err != nil {
			return "", fmt.Errorf("failed to request override : %w", err)
		}
		return "", ErrDeviceMismatch
	}
	return s.startSession(team.ID, deviceID)
}

func (s *AuthService) startSession(teamID int, deviceID string) (string, error) {
	if err := db.DeactivateTeamSessions(teamID); err != nil {
		return "", fmt.Errorf("failed to create session : %w", err)
	}
	sessionToken := utils.TokenGenerator(16)
	session := models.Sessions{
		TeamID:       teamID,
		SessionToken: sessionToken,
		DeviceID:     deviceID,
		Active:       true,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(6 * time.Hour),
	}
	if err := db.CreateSession(session); err != nil {
		return "", fmt.Errorf("failed to create session : %w", err)
	}
	return sessionToken, nil
}

func (s *AuthService) AdminLogin(req models.AdminLoginRequest) (string, error) {
	if s.manifest == nil ||
		req.Username != s.manifest.AdminUsername ||
		req.LoginCode != s.manifest.AdminLoginCode {
		return "", ErrInvalidCredentials
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
		return "", fmt.Errorf("failed to create session : %w", err)
	}
	return token, nil
}
