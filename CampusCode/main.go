package main

import (
	"campuscode/api"
	"campuscode/db"
	"campuscode/judgeEngine"
	
)

func main(){
	//Simulated submission
	judgeEngine.InitializeRunTime()
	db.InitDB()
	judgeEngine.InitializeQueue()
	judgeEngine.StartWorkers(1)
	api.Server()
}