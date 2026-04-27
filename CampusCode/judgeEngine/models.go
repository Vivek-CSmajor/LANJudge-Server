package judgeEngine

import (
	//"gorm.io/gorm"

)


// type Problem struct{
// 	ID int `json:"id" gorm:"primaryKey"`
// 	Title string `json:"title"`
// 	Description string `json:"description"`
// 	Input string `json:"input"`
// 	Output string `json:"output"`
// }
type Submission struct{
	ID int `json:"id" gorm:"primaryKey"`
	UserID int `json:"user_id"`
	ProblemID int `json:"problem_id"`
	Code string `json:"code"`
	Status string `json:"status"`
}

type SubmitRequest  struct { 
	UserID int `json:"user_id"`
	ProblemID int `json:"problem_id"`
	Code string `json:"code"`
	//Input string `json:"input"`
}

type SubmitResponse struct {
	SubmissionID int `json:"submission_id"`
	Status string `json:"status"`
}

