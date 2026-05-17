package main

import (
	"campuscode/api"
	"campuscode/db"
	"campuscode/judge"
)

func main(){
	//Simulated submission
	judge.InitializeRunTime()
	db.InitDB()
	judge.InitializeQueue()
	judge.StartWorkers(3)
	api.Server()
}