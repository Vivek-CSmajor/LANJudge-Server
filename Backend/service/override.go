package service

import (
	"lanjudge/db"
	"lanjudge/models"
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
	if !overrideRecord.IsPending() {
		return models.DeviceOverride{}, ErrAlreadyResolved
	}
	status, ok := models.ResolveOverrideAction(action)
	if !ok {
		return models.DeviceOverride{}, ErrInvalidInput
	}
	overrideRecord.Status = status
	if status == models.OverrideApproved {
		if err := db.UpdateTeamBoundDevice(overrideRecord.TeamID, overrideRecord.NewDeviceID); err != nil {
			return models.DeviceOverride{}, err
		}
		if err := db.DeactivateTeamSessions(overrideRecord.TeamID); err != nil {
			return models.DeviceOverride{}, err
		}
	}
	if err := db.UpdateOverrideStatus(id, overrideRecord.Status); err != nil {
		return models.DeviceOverride{}, err
	}
	return overrideRecord, nil
}
