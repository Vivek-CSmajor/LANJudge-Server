package api

import (
	"net/http"

	"campuscode/api/handlers"
)

func Server() {
	//http.HandleFunc("/submissions/{id}")
	http.HandleFunc("/submit", handlers.SubmitHandler) // Works

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler) //Works
	http.HandleFunc("/login", handlers.LoginHandler)       //Works
	//http.HandleFunc("/logout", handlers.LogoutHandler)
	//http.HandleFunc("/admin/override/{id}", handlers.AdminOverrideHandler)

	//http.HandleFunc("/problems", handlers.ProblemsHandler)//get all problems
	//http.HandleFunc("/problems/{id}", handlers.ProblemDetailHandler)

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) // contest metadata

	//scoreboard
	http.HandleFunc("/scoreboard", handlers.ScoreboardHandler)

	http.ListenAndServe(":8080", http.DefaultServeMux)
}
