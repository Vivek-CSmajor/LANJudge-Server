package contest

import (
	"lanjudge/models"
	"sync"
	"time"
)

var Active *State

// to populate thios in the first place
func NewState(initial models.Contest) *State {
	return &State{contest: initial}
}

// 1. to wrap contest global var
type State struct {
	mu      sync.RWMutex   //3. added mutex cuz touched by more than 1 go routine
	contest models.Contest //2.lower case to unexport the field.
}

func (s *State) Current() models.Contest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.contest
}

func (s *State) Apply(fn func(*models.Contest)) models.Contest {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.contest)
	return s.contest
}

func (s *State) IsOpen(now time.Time) bool {
	c := s.Current()
	return c.Status == models.ContestActive && now.After(c.StartTime) && now.Before(c.EndTime)
}
