package judge

import (
	"campuscode/api/ws"
	"campuscode/db"
	"campuscode/models"
	"log"
)

var SubmissionQueue chan int

func InitializeQueue() {
	SubmissionQueue = make(chan int, 100) // reinitialize the channel
}

func StartWorkers(numOfWorkers int) {
	for i := 0; i < numOfWorkers; i++ {
		go workerLoop(i)
	}
}

func workerLoop(workerID int) {
	for submissionID := range SubmissionQueue {
		if err := db.UpdateSubmissionStatus(submissionID, models.SubmissionRunning); err != nil {
			log.Printf("failed to mark submission %d running: %v", submissionID, err)
			continue
		}
		sub, err := db.GetSubmissionByID(submissionID)
		if err != nil {
			log.Printf("faield to fetch submission : %v ", err)
			continue
		}
		result, err := processSubmission(sub)
		if err != nil {
			log.Printf("Error processing submission %d: %v", submissionID, err)
			if e := db.UpdateSubmissionVerdict(submissionID, models.RuntimeError); e != nil {
				log.Printf("failed to record runtime error verdict: %v", e)
			}
			continue
		}
		sub.Verdict = models.SubmissionVerdict(result)
		if err = db.UpdateSubmissionVerdict(submissionID, sub.Verdict); err != nil {
			log.Printf("failed to update verdict for submission %d: %v", submissionID, err)
		}
		db.UpdateTeamProblemStatsAndScoreboard(sub, models.ActiveContest.StartTime)
		go ws.BroadcastScoreboard()
		if err := db.UpdateSubmissionStatus(submissionID, models.SubmissionDone); err != nil {
			log.Printf("failed to mark submission %d done : %v", submissionID, err)
		}
	}
}
