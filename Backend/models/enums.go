package models

type SubmissionVerdict string

const (
	Accepted          SubmissionVerdict = "AC"
	WrongAnswer       SubmissionVerdict = "WA"
	TimeLimitExceeded SubmissionVerdict = "TLE"
	RuntimeError      SubmissionVerdict = "RE"
	CompilationError  SubmissionVerdict = "CE"
	InternalError     SubmissionVerdict = "IE"
)

type SubmissionStatus string

const (
	SubmissionQueued  SubmissionStatus = "Queued"  // accepted, waiting for a worker
	SubmissionRunning SubmissionStatus = "Running" // a worker has picked it up
	SubmissionDone    SubmissionStatus = "Done"    // judged; Verdict is now set
)

const (
	TeamEnabled      = "enabled"
	TeamDisqualified = "disqualified"
	TeamDisabled     = "disabled"
)

func (t Teams) CanSubmit() (bool, string) {
	if t.Status == nil {
		return true, ""
	}
	switch *t.Status {
	case TeamDisqualified:
		return false, "team is disqualified"
	case TeamDisabled:
		return false, "team account is disabled"
	}
	return true, ""
}

// IsDisqualified reports whether a team has been permanently disqualified —
// narrower than CanAuthenticate, for places where a merely-disabled team
// should still be let through (e.g. logging in).
func (t Teams) IsDisqualified() bool {
	return t.Status != nil && *t.Status == TeamDisqualified
}

type ContestStatus string

const (
	ContestPending ContestStatus = "pending"
	ContestActive  ContestStatus = "active"
	ContestPaused  ContestStatus = "paused"
	ContestEnded   ContestStatus = "ended"
)
