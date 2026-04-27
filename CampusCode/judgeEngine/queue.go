package judgeEngine

import "log"

var submissionQueue chan int

func InitializeQueue(){
	submissionQueue = make(chan int, 100) // reinitialize the channel
}

func StartWorkers(numOfWorkers int){
	for i:=0; i<numOfWorkers; i++{
		go workerLoop(i)
	}
}

func workerLoop(workerID int){
	for submissionID := range submissionQueue{
		log.Printf("worker %d processing submission of id %d",workerID,submissionID)
		processSubmission(getSubmissionByID(submissionID)) //worker processes a job
	}
}

