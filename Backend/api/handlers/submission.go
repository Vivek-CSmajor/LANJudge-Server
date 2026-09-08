package handlers

import (
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { //enforces post only for submission
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	team := TeamFromContext(r)
	req.TeamID = team.ID
	submissionID, err := db.SaveSubmissionToDB(req) // Implement this function to save to DB and return ID
	if err != nil {
		http.Error(w, "db save errro", http.StatusInternalServerError)
		return
	}
	//store code in fs
	submissionDir := judge.SubmissionPathByID(submissionID)
	err = os.MkdirAll(submissionDir, 0755)
	if err != nil {
		http.Error(w, "Failed to create submission directory", http.StatusInternalServerError)
		return
	}
	spec := judge.LanguageSpecs[req.Language]
	submissionPath := submissionDir + "/" + spec.SourceFile
	err = os.WriteFile(submissionPath, []byte(req.Code), 0644)
	if err != nil {
		http.Error(w, "Failed to save submission code", http.StatusInternalServerError)
		return
	}
	// push to queue
	if judge.SubmissionQueue == nil { //exists but aint containing nuthin
		http.Error(w, "queue not initialized", 500)
		return
	}
	judge.SubmissionQueue <- submissionID

	utils.WriteJSON(w, 200, models.SubmitResponse{
		SubmissionID: submissionID,
		Status:       "Queued",
	})
}

// SubmissionDetail embeds models.Submission so its fields (id, team_id,
// problem_id, language, status, verdict, submitted_at) serialize at the same
// level as Code — the frontend sees one flat object, not a nested one.
type SubmissionDetail struct {
	models.Submission
	Code string `json:"code"`
}

func GetSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	//id := strings.TrimPrefix(r.URL.Path,"/submissions/")
	id := r.PathValue("id")
	submission, err := db.GetSubmissionByRawID(id)
	if err != nil {
		http.Error(w, "Submission Not found", http.StatusNotFound)
		return
	}
	team := TeamFromContext(r)
	if submission.TeamID != team.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	code := ""
	if spec, ok := judge.LanguageSpecs[submission.Language]; ok {
		codePath := filepath.Join(judge.SubmissionPathByID(submission.ID), spec.SourceFile)
		if b, err := os.ReadFile(codePath); err == nil {
			code = string(b)
		}
	}

	utils.WriteJSON(w, 200, SubmissionDetail{Submission: submission, Code: code})
}

func ListMySubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	team := TeamFromContext(r)

	submissions, err := db.GetSubmissionsForTeam(team.ID, r.URL.Query().Get("problem_id"))
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, 200, submissions)
}
