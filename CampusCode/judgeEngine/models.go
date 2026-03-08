package judgeEngine

import (
	//"gorm.io/gorm"

)


type Problem struct{
	ID int `json:"id" gorm:"primaryKey"`
	Title string `json:"title"`
	Description string `json:"description"`
	Input string `json:"input"`
	Output string `json:"output"`
}
type Submission struct{
	ID int `json:"id" gorm:"primaryKey"`
	ProblemID int `json:"problem_id"`
	Code string `json:"code"`
	Result string `json:"result"`
}
