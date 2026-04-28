package main

import (
	"campuscode/judgeEngine"
	"net/http"
)

func main(){
	// Simulated submission
	judgeEngine.InitializeRunTime()
	judgeEngine.InitDB()
	judgeEngine.InitializeQueue()
	judgeEngine.StartWorkers(1)
	http.HandleFunc("/submit", judgeEngine.SubmitHandler)
	http.ListenAndServe(":8080", http.DefaultServeMux)
}