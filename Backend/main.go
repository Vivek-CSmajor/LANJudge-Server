package main

import (
	"campuscode/api"
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
	"log"
)

func main(){
	//Simulated submission
	judge.InitializeRunTime()
	db.InitDB()
	models.ActiveContest = db.LoadOrInitContest()
	manifest,err := judge.LoadManifest()
	if err != nil {
		log.Fatal("failed to load manifest: ",err)
	}
	judge.ActiveManifest = manifest
	judge.InitializeQueue()
	judge.StartWorkers(3)
	api.Server()
}