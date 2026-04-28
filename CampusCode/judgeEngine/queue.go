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
		DB.Model(&Submission{}).
		Where("id = ?",submissionID).
		Update("status", "Running")
		log.Printf("worker %d processing submission of id %d",workerID,submissionID)
		sub, err  := getSubmissionByID(submissionID)
		result , err :=processSubmission(sub) //worker processes a job
		if err != nil {
			log.Printf("Error processing submission %d: %v", submissionID, err)
			DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", RuntimeError)
		}

		if result == string(Accepted){
			DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", Accepted)
		}else if result == string(WrongAnswer){
			DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", WrongAnswer)
		}else if result == string(CompilationError){
			DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", CompilationError)
		}else if result == string(TimeLimitExceeded){
			DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", TimeLimitExceeded)
		}else if result == string(RuntimeError){
    		DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("verdict", RuntimeError)
		}
		DB.Model(&Submission{}).
			Where("id = ?",submissionID).
			Update("status", "Done")
	}
}

