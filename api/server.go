package api

import (
	"net/http"

	"campuscode/api/handlers"
)

func Server() {
	http.HandleFunc("/submit", handlers.SubmitHandler) // Works

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler) //Works
	http.HandleFunc("/login", handlers.LoginHandler)       //Works
	//http.HandleFunc("/logout", handlers.LogoutHandler)
	//http.HandleFunc("/admin/override/{id}", handlers.AdminOverrideHandler)

	http.HandleFunc("/problems", handlers.ProblemsHandler)//get all problems returns only meta data not the whole problem, works as problem navigator
	http.HandleFunc("/problems/{id}", handlers.ProblemHandler)

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) // contest metadata

	//scoreboard
	http.HandleFunc("/scoreboard", handlers.ScoreboardHandler)//works

	//admin
	http.HandleFunc("/admin/submissions",handlers.AdminGetSubmissionsHandler)
	http.HandleFunc("/admin/submissions/{id}",handlers.AdminGetSubmissionHandler)
	http.HandleFunc("/admin/teams/{id}",handlers.AdminChangeTeamStatusHandler)
	http.ListenAndServe(":8080", http.DefaultServeMux)
}
