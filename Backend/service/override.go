package service

import (
	"campuscode/db"
	"campuscode/models"
)

type OverrideService struct{}

func NewOverrideService() *OverrideService {
	return &OverrideService{}
}

func (s *OverrideService) List() ([]models.DeviceOverride, error) {
	return db.GetAllOverrides()
}

func (s *OverrideService) Approve(id int, action string) (models.DeviceOverride, error) {
	overrideRecord, err := db.GetOverrideByID(id)
	if err != nil {
		return models.DeviceOverride{}, ErrNotFound
	}
	if overrideRecord.Status != "pending" {
		return models.DeviceOverride{}, ErrAlreadyResolved
	}
	if action == "approve" {
		overrideRecord.Status = "approved"
		if err := db.UpdateTeamBoundDevice(overrideRecord.TeamID, overrideRecord.NewDeviceID); err != nil {
			return models.DeviceOverride{}, err
		}
		if err := db.DeactivateTeamSessions(overrideRecord.TeamID); err != nil {
			return models.DeviceOverride{}, err
		}
	} else if action == "reject" {
		overrideRecord.Status = "rejected"
	}
	if err := db.UpdateOverrideStatus(id, overrideRecord.Status); err != nil {
		return models.DeviceOverride{}, err
	}
	return overrideRecord, nil
}
