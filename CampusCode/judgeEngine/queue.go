package judgeEngine

import (
	"log"
	"campuscode/db"
	"campuscode/models"
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
	for submissionID := range SubmissionQueue{
		db.DB.Model(&models.Submission{}).
		Where("id = ?",submissionID).
		Update("status", "Running")
		log.Printf("worker %d processing submission of id %d",workerID,submissionID)
		sub, err  := db.GetSubmissionByID(submissionID)
		result , err :=processSubmission(sub) //worker processes a job
		if err != nil {
			log.Printf("Error processing submission %d: %v", submissionID, err)
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.RuntimeError)
		}

		if result == string(models.Accepted){
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.Accepted)
		}else if result == string(models.WrongAnswer){
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.WrongAnswer)
		}else if result == string(models.CompilationError){
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.CompilationError)
		}else if result == string(models.TimeLimitExceeded){
			db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.TimeLimitExceeded)
		}else if result == string(models.RuntimeError){
    		db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", models.RuntimeError)
		}
		db.DB.Model(&models.Submission{}).
			Where("id = ?",submissionID).
			Update("status", "Done")
	}
}

