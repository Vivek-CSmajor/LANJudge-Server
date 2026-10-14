package main

import (
	"lanjudge/api"
	"lanjudge/api/handlers"
	"lanjudge/contest"
	"lanjudge/db"
	"lanjudge/judge"
	"fmt"
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
	if err := chdirToExecutableDir(); err != nil {
		log.Fatal("failed to determine the server's own directory: ", err)
	}
	if err := ensureContestData(); err != nil {
		fmt.Fprintln(os.Stderr, "\n"+err.Error()+"\n")
		os.Exit(1)
	}

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

	const dashboardURL = "http://localhost:8080/"
	log.Println("Admin dashboard will be available at:", dashboardURL)
	go openDashboardWhenReady(dashboardURL)

	api.Server(h)
}
