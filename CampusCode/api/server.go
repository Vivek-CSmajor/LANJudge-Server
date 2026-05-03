package api

import (
	"net/http"
	
	"campuscode/api/handlers"
)

func Server(){
	http.HandleFunc("/submit", handlers.SubmitHandler) // Works

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler) //done .IDK bout working	
	http.HandleFunc("/login", handlers.LoginHandler)//done .IDK bout working
	//http.HandleFunc("/logout", handlers.LogoutHandler) 
	//http.HandleFunc("/admin/override/{id}", handlers.AdminOverrideHandler)
	

	//http.HandleFunc("/problems", handlers.ProblemsHandler)
	//http.HandleFunc("/problems/{id}", handlers.ProblemDetailHandler)

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) // contest metadata


	http.ListenAndServe(":8080", http.DefaultServeMux)
}