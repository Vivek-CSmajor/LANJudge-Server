package db

import (

	"campuscode/models"
	"log"
	"time"
	//"gorm.io/gorm"
)

func SaveSubmissionToDB(sub models.SubmitRequest) int{ // saves submission data to db and return the subID
	submission := models.Submission{
		TeamID: sub.TeamID,
		ProblemID: sub.ProblemID,
		//Code: sub.Code,
		Status : "Queued",	
		SubmittedAt: time.Now(),
	}
	DB.Create(&submission)
	return submission.ID
}

func GetSubmissionByID(subID int) (models.Submission, error){
	var sub models.Submission
	err := DB.First(&sub, subID).Error
	if err != nil {
		return models.Submission{},err
	}
	return sub, nil
}

//i write like this as my first language was c# and it folllows this repository thing very strictly
func RegisterTeamRepository(req models.RegisterResponse){
	registration := models.Teams{
		Name : req.Name,
		LoginCode: req.LoginCode,
		BoundDeviceID: nil,
	}
	DB.Create(&registration)
}

func UpdateTeamProblemStatsAndScoreboard(sub models.Submission,startTime time.Time){
	var stats models.TeamProblemStats
	DB.FirstOrCreate(&stats, models.TeamProblemStats{
		TeamID: sub.TeamID,
		ProblemID: sub.ProblemID,
		WrongAttemptCount: 0,
		Solved: false,
		FirstAcTime: nil,
		Penalty: nil,
	})
	var team models.ScoreBoard
	DB.FirstOrCreate(&team, models.ScoreBoard{
		TeamID: stats.TeamID,
	})
	//case 2 already solved
	if(stats.Solved == true){
		return
	}

	//case 3 if AC verdict
	if(sub.Verdict == models.Accepted){
		stats.Solved = true
		stats.FirstAcTime = &sub.SubmittedAt//bit of confusion, if a value is *int u must send it address of the value u wanna assign it not the value.
		minutesFromStart := int(sub.SubmittedAt.Sub(startTime).Minutes())
		penalty := minutesFromStart + (20*stats.WrongAttemptCount)
		stats.Penalty = &penalty
		team.TotalSolved++
		team.TotalPenalty += penalty
		team.LastAcTime = sub.SubmittedAt
	}

	if(sub.Verdict == models.WrongAnswer || sub.Verdict == models.RuntimeError){
		stats.WrongAttemptCount++
	}
	res := DB.Save(&stats)
	log.Println(res.Error)
	log.Println(res.RowsAffected)
	
	result := DB.Save(&team)

	log.Println(result.Error)
	log.Println(result.RowsAffected)

}
