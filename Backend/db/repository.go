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
		Language: sub.Language,
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
func RegisterTeamRepository(req models.RegisterResponse)error{
	registration := models.Teams{
		Name : req.Name,
		LoginCode: req.LoginCode,
		BoundDeviceID: nil,
	}
	err := DB.Create(&registration).Error
	if err !=nil {
		return err
	}
	return nil
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
	var t models.Teams
	DB.First(&t,stats.TeamID)
	DB.FirstOrCreate(&team, models.ScoreBoard{
		TeamID: stats.TeamID,
	})
	team.TeamName = t.Name
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


func GetAllProblems()([]models.Problem, error){
	var problems []models.Problem
	err := DB.Order("id ASC").Find(&problems).Error
	if err != nil {
		return nil,  err 
	}
	return problems, nil
}

func LoadOrInitContest() *models.Contest {
	var contest models.Contest
	err := DB.First(&contest).Error
	if err != nil {
		contest = models.Contest{
			Title:     "Contest",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Status:    models.ContestPending,
		}
		DB.Create(&contest)
	}
	return &contest
}

func GetTeamFromSubmission(token string)(models.Teams,error){
	var team models.Teams
	var session models.Sessions
	err := DB.Where("session_token = ?",token).Find(&session).Error
	if err != nil{
		return models.Teams{}, err
	}
	err = DB.First(&team,session.TeamID).Error
	if err != nil {
		return models.Teams{},err
	}
	return team,err
}