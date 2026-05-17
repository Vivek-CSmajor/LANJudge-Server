package judge

import (
	//"fmt"
	"os"
)
const basePath  = "server_data"

func InitializeRunTime()error{
	// create runtime directory if it doesn't exist
	fs := FileSystem{
		BasePath: basePath,
		ContestPath: basePath + "/contest",
		Runtime: basePath + "/runtime",
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



// These exist to make parsing JSON recieved from the cloud backend to be deserialized
type Manifest struct {
	//ContestID string `json:"contest_id"`
	Name string `json:"name"`
	DurationMinutes int `json:"duration_minutes"`
	Problems []Problem `json:"problems"`
}

type Problem struct {
	ID string `json:"id"`
	Path string `json:"path"`
}

type ProblemMeta struct {
	TimeLimitMS int `json:"time_limit_ms"`
	MemoryLimitMB int `json:"memory_limit_mb"`
}

type FileSystem struct {
	BasePath string 
	ContestPath string 
	Runtime string
}




