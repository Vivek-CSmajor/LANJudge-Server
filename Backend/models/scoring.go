package models

const WrongAttemptPenaltyMinutes = 20

func Penalty(minutesFromStart, wrongAttempts int) int {
	return minutesFromStart + (WrongAttemptPenaltyMinutes * wrongAttempts)
}
