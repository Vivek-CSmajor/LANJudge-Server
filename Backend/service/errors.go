package service

import "fmt"

var (
	ErrConflict           = fmt.Errorf("conflict")
	ErrInvalidInput       = fmt.Errorf("invalid input")
	ErrAlreadyResolved    = fmt.Errorf("already resolved")
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrDisqualified       = fmt.Errorf("team is disqualified")
	ErrDeviceMismatch     = fmt.Errorf("device mismatch")
	ErrBusy               = fmt.Errorf("server busy")
)
