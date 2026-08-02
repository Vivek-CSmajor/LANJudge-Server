package main

import (
	"campuscode/api"
	"campuscode/db"
	"campuscode/judge"
	"campuscode/models"
)

func main(){
	//Simulated submission
	judge.InitializeRunTime()
	db.InitDB()
	models.ActiveContest = db.LoadOrInitContest()
	judge.InitializeQueue()
	judge.StartWorkers(3)
	api.Server()
}