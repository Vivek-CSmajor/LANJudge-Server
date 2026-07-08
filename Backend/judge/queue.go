package judge

import (
	"campuscode/api/ws"
	"campuscode/db"
	"campuscode/models"
	"log"
)

var SubmissionQueue chan int

func InitializeQueue(){
	SubmissionQueue = make(chan int, 100) // reinitialize the channel
}
	
func StartWorkers(numOfWorkers int){
for i:=0; i<numOfWorkers; i++{
		go workerLoop(i)
	}
}

func workerLoop(workerID int){
	var contest models.Contest
	db.DB.First(&contest)
	for submissionID := range SubmissionQueue{
		db.DB.Model(&models.Submission{}).
		Where("id = ?",submissionID).
		Update("status", "Running")
		log.Printf("worker %d processing submission of id %d",workerID,submissionID)
		sub, err  := db.GetSubmissionByID(submissionID)
		if err != nil{
			log.Printf("faield to fetch submission : %v ", err)
			continue
		}
		result , err :=processSubmission(sub) //worker processes a job
		if err != nil {
			log.Printf("Error processing submission %d: %v", submissionID, err)
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.RuntimeError)
			sub.Verdict = models.SubmissionVerdict(result)
		}
		db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", result)
			sub.Verdict = models.SubmissionVerdict(result)
			db.UpdateTeamProblemStatsAndScoreboard(sub,contest.StartTime)
			ws.BroadcastScoreboard()
		db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("status", "Done")
	}
}

