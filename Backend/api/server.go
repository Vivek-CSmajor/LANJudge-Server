package api

import (
	"net/http"

	"campuscode/api/handlers"
	"campuscode/api/ws"
)

func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Device-ID,X-Session-Token")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func Server() {
	http.HandleFunc("/submit", handlers.SubmitHandler)
	http.HandleFunc("/submission/{id}",handlers.GetSubmissionHandler)

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler) 
	http.HandleFunc("/login", handlers.LoginHandler)       
	//http.HandleFunc("/logout", handlers.LogoutHandler)
	//http.HandleFunc("/admin/override/{id}", handlers.AdminOverrideHandler)

	http.HandleFunc("/problems", handlers.ProblemsHandler)
	http.HandleFunc("/problems/{id}", handlers.ProblemHandler)

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) 

	//scoreboard
	http.HandleFunc("/scoreboard", handlers.ScoreboardHandler)

	//admin
	http.HandleFunc("/admin/submissions",handlers.AdminGetSubmissionsHandler)
	http.HandleFunc("/admin/submissions/{id}", handlers.AdminGetSubmissionHandler)
	http.HandleFunc("/admin/teams/{id}", handlers.AdminChangeTeamStatusHandler)
	http.HandleFunc("/admin/teams",handlers.AdminGetTeamsHandler)
	http.HandleFunc("/admin/scoreboard",handlers.AdminGetScoreboard)


	http.HandleFunc("/admin/contest",handlers.AdminChangeContestState)

	ws.StartTimeSyncTicker()

	http.ListenAndServe(":8080", loggingMiddleware(cors(http.DefaultServeMux)))
}
