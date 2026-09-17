package judge

import (
	"campuscode/db"
	"campuscode/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const basePath = "server_data"
const contestPath = basePath + "/contest"
const runtimePath = basePath + "/runtime"

func InitializeRunTime() error {
	// create runtime directory if it doesn't exist
	fs := FileSystem{
		BasePath:    basePath,
		ContestPath: contestPath,
		Runtime:     runtimePath,
	}
	paths := []string{
		fs.Runtime,
		fs.Runtime + "/submissions",
		fs.Runtime + "/workspaces",
	}
	for _, path := range paths { // 0755 is rwxr-xr-x
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func WorkspacePathByID(id int) string {
	path := filepath.Join(runtimePath, "workspaces", strconv.Itoa(id))
	return path
}

func SubmissionPathByID(id int) string {
	path := filepath.Join(runtimePath, "submissions", strconv.Itoa(id))
	return path
}

func TestsDir(id int) string { //asdasdasd
	path := filepath.Join(contestPath, "problems", strconv.Itoa(id), "tests")
	return path
}

func ProblemMetaPath(id int) string {
	return filepath.Join(contestPath, "problems", strconv.Itoa(id), "meta.json")
}

func LoadProblemMeta(id int) (ProblemMeta, error) {
	data, err := os.ReadFile(ProblemMetaPath(id))
	if err != nil {
		return ProblemMeta{}, err
	}
	var meta ProblemMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return ProblemMeta{}, err
	}
	return meta, nil
}

func LoadManifest() (*Manifest, error) {
	mnf, err := os.ReadFile(filepath.Join(contestPath, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(mnf, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// These exist to make parsing JSON recieved from the cloud backend to be deserialized
type Manifest struct {
	ContestID       string    `json:"contest_id"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"duration_minutes"`
	Problems        []Problem `json:"problems"`
	AdminUsername   string    `json:"admin_user_name"`
	AdminLoginCode  string    `json:"admin_login_code"`
}

type Problem struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
}

func LoadOrInitProblems(manifest *Manifest) error {
	for _, p := range manifest.Problems {
		id, err := strconv.Atoi(p.ID)
		if err != nil {
			return fmt.Errorf("invalid problem id %q in manifest: %w", p.ID, err)
		}
		exists, err := db.ExistsByID[models.Problem](id)
		if err != nil {
			return fmt.Errorf("failed to check problem %d: %w", id, err)
		}
		if exists {
			continue
		}
		meta, err := LoadProblemMeta(id)
		if err != nil {
			return fmt.Errorf("failed to load meta.json for problem %s: %w", p.ID, err)
		}

		problem := models.Problem{
			ID:          id,
			Title:       p.Title,
			TimeLimit:   time.Duration(meta.TimeLimitMS) * time.Millisecond,
			MemoryLimit: meta.MemoryLimitMB,
		}
		err = db.CreateProblem(problem)
		if err != nil {
			return fmt.Errorf("failed to insert problem %d: %w", id, err)
		}
	}
	return nil
}

type ProblemMeta struct {
	TimeLimitMS   int `json:"time_limit_ms"`
	MemoryLimitMB int `json:"memory_limit_mb"`
}

type FileSystem struct {
	BasePath    string
	ContestPath string
	Runtime     string
}
