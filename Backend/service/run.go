package service

import (
	"lanjudge/judge"
)

type RunService struct {
	semaphore chan struct{}
}

func NewRunService(semaphore chan struct{}) *RunService {
	return &RunService{semaphore: semaphore}
}

func (s *RunService) Run(language, code, input string) (string, string, error) {
	///gotta use the semaphore here since it acts liek a buffered channel with a limit. case 1 sends empty struct (since we dont need data)
	/// defer should be written right next to it so no matter what happens if (the logic in between if u write it) succeeds or not, defer always runs
	// suppose if the logic in bw panics then use semaphore accumulates overtime with ghost values that never free.
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
		output, errMsg := judge.RunAdHoc(language, code, input)
		return output, errMsg, nil
	default:
		return "", "", ErrBusy
	}
}
