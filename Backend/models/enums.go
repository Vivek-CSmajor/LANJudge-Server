package models

type SubmissionVerdict string
const(
	Accepted SubmissionVerdict = "AC"
	WrongAnswer SubmissionVerdict = "WA"
	TimeLimitExceeded SubmissionVerdict = "TLE"
	RuntimeError SubmissionVerdict = "RE"
	CompilationError SubmissionVerdict = "CE"
	InternalError SubmissionVerdict = "IE"
)

type ContestStatus string
const (
	ContestPending ContestStatus = "pending"
	ContestActive  ContestStatus = "active"
	ContestPaused  ContestStatus = "paused"
	ContestEnded   ContestStatus = "ended"
)