package api

import (
	"net/http"

	"campuscode/api/handlers"
)

func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Device-ID")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func Server() {
	http.HandleFunc("/submit", handlers.SubmitHandler) // Works

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler) //Works
	http.HandleFunc("/login", handlers.LoginHandler)       //Works
	//http.HandleFunc("/logout", handlers.LogoutHandler)
	//http.HandleFunc("/admin/override/{id}", handlers.AdminOverrideHandler)

	http.HandleFunc("/problems", handlers.ProblemsHandler)
	http.HandleFunc("/problems/{id}", handlers.ProblemHandler)

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) // contest metadata

	//scoreboard
	http.HandleFunc("/scoreboard", handlers.ScoreboardHandler)

	//admin
	http.HandleFunc("/admin/submissions",handlers.AdminGetSubmissionsHandler)
	http.HandleFunc("/admin/submissions/{id}", handlers.AdminGetSubmissionHandler)
	http.HandleFunc("/admin/teams/{id}", handlers.AdminChangeTeamStatusHandler)
	http.ListenAndServe(":8080", cors(http.DefaultServeMux))
}
