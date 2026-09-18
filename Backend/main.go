package main

import (
	"campuscode/api"
	"campuscode/api/handlers"
	"campuscode/contest"
	"campuscode/db"
	"campuscode/judge"
	"log"
	"os"
	"strconv"
)

func judgeWorkerCount() int {
	if raw := os.Getenv("JUDGE_WORKERS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return 8
}

func main() {
	//Simulated submission
	if err := judge.InitializeRunTime(); err != nil {
		log.Fatal("failed to initialize runtime directories: ", err)
	}
	db.InitDB()
	initialContest := db.LoadOrInitContest()
	contest.Active = contest.NewState(*initialContest)
	manifest, err := judge.LoadManifest()
	if err != nil {
		log.Fatal("failed to load manifest: ", err)
	}
	if err := judge.LoadOrInitProblems(manifest); err != nil {
		log.Fatal("failed to load problems: ", err)
	}
	queue := judge.InitializeQueue()
	judge.StartWorkers(judgeWorkerCount(), queue, contest.Active)
	runSemaphore := make(chan struct{}, 3)
	h := handlers.NewHandlers(contest.Active, manifest, queue, runSemaphore)
	api.Server(h)
}
