package utils

import (
	"lanjudge/models"
	"crypto/rand"
	"encoding/json"
	"sync"
	"time"

	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func TokenGenerator(n int) string { //8 bit for logincode and more for sessionToken
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	b := make([]byte, n)
	rand.Read(b)

	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[num.Int64()]
		//b[i] = alphabet[int(b[i])%32]
	}
	log.Print(string(b))
	return string(b)
}

func CopyFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}

func ListInputFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries { //range returns index and value
		name := entry.Name()
		if strings.HasSuffix(name, ".in") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files, nil
}

func WriteJSON(w http.ResponseWriter, status int, a any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(a)
}

func NormalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n") //convert windwos line endings to unix
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func ReadProblemStatement(problemPath string) (string, error) {
	path := filepath.Join(problemPath, "statement.md")
	data, err := os.ReadFile(path)
	return string(data), err
}

// to populate thios in the first place
func NewState(initial models.Contest) *State {
	return &State{contest: initial}
}

// 1. to wrap contest global var
type State struct {
	mu      sync.RWMutex   //3. added mutex cuz touched by more than 1 go routine
	contest models.Contest //2.lower case to unexport the field.
}

func (s *State) Current() models.Contest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.contest
}

func (s *State) Apply(fn func(*models.Contest)) models.Contest {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.contest)
	return s.contest
}

func (s *State) IsOpen(now time.Time) bool {
	c := s.Current()
	return c.Status == models.ContestActive && now.After(c.StartTime) && now.Before(c.EndTime)
}
