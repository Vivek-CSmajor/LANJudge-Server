package main

import (
	"campuscode/api"
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"log"
)

func main() {
	//Simulated submission
	if err := judge.InitializeRunTime(); err != nil {
		log.Fatal("failed to initialize runtime directories: ", err)
	}
	db.InitDB()
	models.ActiveContest = db.LoadOrInitContest()
	manifest, err := judge.LoadManifest()
	if err != nil {
		log.Fatal("failed to load manifest: ", err)
	}
	judge.ActiveManifest = manifest
	if err := judge.LoadOrInitProblems(manifest); err != nil {
		log.Fatal("failed to load problems: ", err)
	}
	judge.InitializeQueue()
	judge.StartWorkers(3)
	api.Server()
}
