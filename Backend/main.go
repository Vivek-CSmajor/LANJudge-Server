package main

import (
	"campuscode/api"
	"campuscode/api/handlers"
	"campuscode/contest"
	"campuscode/db"
	"campuscode/judge"
	"log"
)

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
	judge.StartWorkers(3, queue, contest.Active)
	runSemaphore := make(chan struct{}, 3)
	h := handlers.NewHandlers(contest.Active, manifest, queue, runSemaphore)
	api.Server(h)
}
