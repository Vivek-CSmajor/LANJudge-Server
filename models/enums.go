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

