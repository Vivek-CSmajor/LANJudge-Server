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

type ContestStatus string

const (
	ContestPending ContestStatus = "pending"
	ContestActive  ContestStatus = "active"
	ContestPaused  ContestStatus = "paused"
	ContestEnded   ContestStatus = "ended"
)
