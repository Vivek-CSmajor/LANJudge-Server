package db

import (
	"campuscode/models"
	"log"
	"time"
)

func CreateModel[T any](model *T) error {
	return conn.Create(model).Error
}

func SaveModel[T any](model *T) error {
	return conn.Save(model).Error
}

func GetByID[T any](id int) (T, error) {
	var row T
	if err := conn.First(&row, id).Error; err != nil {
		var zero T
		return zero, err
	}
	return row, nil
}

func GetAll[T any](order string) ([]T, error) {
	var rows []T
	if err := conn.Order(order).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func ExistsByID[T any](id int) (bool, error) {
	var count int64
	err := conn.Model(new(T)).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateFields[T any](fields map[string]any, query string, args ...any) (int64, error) {
	res := conn.Model(new(T)).Where(query, args...).Updates(fields)
	return res.RowsAffected, res.Error
}

func UpdateFieldsByID[T any](id int, fields map[string]any) (int64, error) {
	return UpdateFields[T](fields, "id = ?", id)
}

func GetOne[T any](query string, args ...any) (T, error) {
	var row T
	if err := conn.Where(query, args...).First(&row).Error; err != nil {
		var zero T
		return zero, err
	}
	return row, nil
}

func SaveSubmissionToDB(sub models.SubmitRequest) (int, error) { // saves submission data to db and return the subID
	submission := models.Submission{
		TeamID:    sub.TeamID,
		ProblemID: sub.ProblemID,
		//Code: sub.Code,
		Status:      string(models.SubmissionQueued),
		SubmittedAt: time.Now(),
		Language:    sub.Language,
	}
	if err := CreateModel(&submission); err != nil {
		return 0, err
	}
	return submission.ID, nil
}

func GetTeamByID(id int) (models.Teams, error) {
	return GetByID[models.Teams](id)
}

func GetSubmissionByID(subID int) (models.Submission, error) {
	return GetByID[models.Submission](subID) // works only if the id is a PK
}

func GetSubmissionByRawID(id string) (models.Submission, error) {
	return GetOne[models.Submission]("id = ?", id) //works for any type //had to write this because i was getting id from r.PathValue which gives a string back
}

func GetAllSubmissions() ([]models.Submission, error) {
	return GetAll[models.Submission]("submitted_at DESC")
}

func GetSubmissionsForTeam(teamID int, problemID string) ([]models.Submission, error) {
	query := conn.Where("team_id = ?", teamID)
	if problemID != "" {
		query = query.Where("problem_id = ?", problemID)
	}
	var submissions []models.Submission
	if err := query.Order("submitted_at DESC").Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

func UpdateSubmissionStatus(id int, status models.SubmissionStatus) error {
	_, err := UpdateFieldsByID[models.Submission](id, map[string]any{"status": status})
	return err
}

func UpdateSubmissionVerdict(id int, verdict models.SubmissionVerdict) error {
	_, err := UpdateFieldsByID[models.Submission](id, map[string]any{"verdict": verdict})
	return err
}

func ResetSubmissionForRejudge(id int) error {
	_, err := UpdateFieldsByID[models.Submission](id, map[string]any{
		"status":  models.SubmissionQueued,
		"verdict": "",
	})
	return err
}

// i write like this as my first language was c# and it folllows this repository thing very strictly
func RegisterTeamRepository(req models.RegisterResponse) error {
	registration := models.Teams{
		Name:          req.Name,
		LoginCode:     req.LoginCode,
		BoundDeviceID: nil,
	}
	return CreateModel(&registration)
}

func GetAllTeams() ([]models.Teams, error) {
	return GetAll[models.Teams]("id ASC")
}

func UpdateTeamStatus(teamID int, status string) (int64, error) {
	return UpdateFieldsByID[models.Teams](teamID, map[string]any{"status": status})
}

func UpdateTeamBoundDevice(teamID int, deviceID string) error {
	_, err := UpdateFieldsByID[models.Teams](teamID, map[string]any{"bound_device_id": deviceID})
	return err
}

func GetTeamByNameAndLoginCode(name, loginCode string) (models.Teams, error) {
	return GetOne[models.Teams]("name = ? AND login_code = ?", name, loginCode)
}

func CreateSession(session models.Sessions) error {
	return CreateModel(&session)
}

func DeactivateTeamSessions(teamID int) error {
	_, err := UpdateFields[models.Sessions](
		map[string]any{"active": false},
		"team_id = ?", teamID,
	)
	return err
}

func GetActiveSessionByToken(token string) (models.Sessions, error) {
	return GetOne[models.Sessions]("session_token = ? AND active = ? AND expires_at > ?", token, true, time.Now())
}

func GetActiveAdminSessionByToken(token string) (models.Sessions, error) {
	return GetOne[models.Sessions]("session_token = ? AND active = ? AND admin = ?", token, true, true)
}

func GetAllOverrides() ([]models.DeviceOverride, error) {
	return GetAll[models.DeviceOverride]("created_at DESC")
}

func GetOverrideByID(id int) (models.DeviceOverride, error) {
	return GetByID[models.DeviceOverride](id)
}

func UpdateOverrideStatus(id int, status string) error {
	_, err := UpdateFieldsByID[models.DeviceOverride](id, map[string]any{"status": status})
	return err
}

func CreateDeviceOverride(override models.DeviceOverride) error {
	return CreateModel(&override)
}

func GetAllProblems() ([]models.Problem, error) {
	return GetAll[models.Problem]("id ASC")
}

func GetProblemByID(id string) (models.Problem, error) {
	return GetOne[models.Problem]("id = ?", id)
}

func CreateProblem(problem models.Problem) error {
	return CreateModel(&problem)
}

func SaveContest(contest *models.Contest) error {
	return SaveModel(contest)
}

func LoadOrInitContest() *models.Contest {
	var contest models.Contest
	err := conn.First(&contest).Error
	if err != nil {
		contest = models.Contest{
			Title:     "Contest",
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Status:    models.ContestPending,
		}
		conn.Create(&contest)
	}
	return &contest
}

func UpdateTeamProblemStatsAndScoreboard(sub models.Submission, startTime time.Time) {
	var stats models.TeamProblemStats
	conn.FirstOrCreate(&stats, models.TeamProblemStats{
		TeamID:            sub.TeamID,
		ProblemID:         sub.ProblemID,
		WrongAttemptCount: 0,
		Solved:            false,
		FirstAcTime:       nil,
		Penalty:           nil,
	})
	var team models.ScoreBoard
	var t models.Teams
	conn.First(&t, stats.TeamID)
	conn.FirstOrCreate(&team, models.ScoreBoard{
		TeamID: stats.TeamID,
	})
	team.TeamName = t.Name
	//case 2 already solved
	if stats.Solved == true {
		return
	}

	//case 3 if AC verdict
	if sub.Verdict == models.Accepted {
		stats.Solved = true
		stats.FirstAcTime = &sub.SubmittedAt //bit of confusion, if a value is *int u must send it address of the value u wanna assign it not the value.
		minutesFromStart := int(sub.SubmittedAt.Sub(startTime).Minutes())
		penalty := minutesFromStart + (20 * stats.WrongAttemptCount)
		stats.Penalty = &penalty
		team.TotalSolved++
		team.TotalPenalty += penalty
		team.LastAcTime = sub.SubmittedAt
	}

	if sub.Verdict == models.WrongAnswer || sub.Verdict == models.RuntimeError {
		stats.WrongAttemptCount++
	}
	res := conn.Save(&stats)
	log.Println(res.Error)
	log.Println(res.RowsAffected)

	result := conn.Save(&team)
	log.Println(result.Error)
	log.Println(result.RowsAffected)
}

func GetScoreboard() ([]models.ScoreBoard, error) {
	var scoreboard []models.ScoreBoard
	err := conn.Joins("JOIN teams ON teams.id = score_boards.team_id").
		Where("teams.status IS NULL OR teams.status <> ?", "disqualified").
		Order("total_solved DESC").
		Order("total_penalty ASC").
		Order("last_ac_time ASC").
		Find(&scoreboard).Error
	for i := range scoreboard {
		scoreboard[i].Rank = i + 1
	}
	return scoreboard, err
}
