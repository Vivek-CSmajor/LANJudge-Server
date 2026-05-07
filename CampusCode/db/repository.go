package db

import (
	"campuscode/models"
)

func SaveSubmissionToDB(sub models.SubmitRequest) int{ // saves submission data to db and return the subID
	submission := models.Submission{
		UserID: sub.UserID,
		ProblemID: sub.ProblemID,
		//Code: sub.Code,
		Status : "Queued",	
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

func CreateSessionRepository(req models.LoginRequest){
	
}
