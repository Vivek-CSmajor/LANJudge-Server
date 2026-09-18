package judge

import (
	"campuscode/api/ws"
	"campuscode/contest"
	"campuscode/db"
	"campuscode/models"
	"log"
)

func InitializeQueue() chan int {
	return make(chan int, 100)
}

func StartWorkers(numOfWorkers int, queue chan int, contestState *contest.State) {
	for i := 0; i < numOfWorkers; i++ {
		go workerLoop(i, queue, contestState)
	}
}

func workerLoop(workerID int, queue chan int, contestState *contest.State) {
	for submissionID := range queue {
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

			verdict := models.SubmissionVerdict(result)
			if verdict == "" {
				verdict = models.InternalError  // fallback if result is empty
			}

			if e := db.UpdateSubmissionVerdict(submissionID, verdict); e != nil {
				log.Printf("failed to record verdict: %v", e)
			}
			if e := db.UpdateSubmissionStatus(submissionID, models.SubmissionDone); e != nil {
				log.Printf("failed to mark submission done: %v", e)
			}
			continue
		}
		sub.Verdict = models.SubmissionVerdict(result)
		if err = db.UpdateSubmissionVerdict(submissionID, sub.Verdict); err != nil {
			log.Printf("failed to update verdict for submission %d: %v", submissionID, err)
		}
		db.UpdateTeamProblemStatsAndScoreboard(sub, contestState.Current().StartTime)
		go ws.BroadcastScoreboard()
		if err := db.UpdateSubmissionStatus(submissionID, models.SubmissionDone); err != nil {
			log.Printf("failed to mark submission %d done : %v", submissionID, err)
		}
	}
}
