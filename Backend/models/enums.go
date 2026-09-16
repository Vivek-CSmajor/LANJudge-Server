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

const (
	OverridePending  = "pending"
	OverrideApproved = "approved"
	OverrideRejected = "rejected"
)

func (o DeviceOverride) IsPending() bool {
	return o.Status == OverridePending
}

func ResolveOverrideAction(action string) (status string, ok bool) {
	switch action {
	case "approve":
		return OverrideApproved, true
	case "reject":
		return OverrideRejected, true
	}
	return "", false
}
