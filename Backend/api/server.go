package api

import (
	"log"
	"net/http"
	//"campuscode/api/middleware"
	"campuscode/api/handlers"
	"campuscode/api/ws"
)

func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Device-ID,X-Session-Token,Authorization")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func Server() {
	http.HandleFunc("/health", handlers.HealthHandler)

	http.HandleFunc("/submit", handlers.RequireAuth(handlers.SubmitHandler))
	http.HandleFunc("/submission/{id}",handlers.RequireAuth(handlers.GetSubmissionHandler))
	http.HandleFunc("/run", handlers.RequireAuth(handlers.RunHandler))

	//auth rotues
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	//http.HandleFunc("/logout", handlers.LogoutHandler)

	http.HandleFunc("/problems", handlers.RequireAuth(handlers.ProblemsHandler))
	http.HandleFunc("/problems/{id}", handlers.RequireAuth(handlers.ProblemHandler))

	//contest handlers
	//http.HandleFunc("/contests", handlers.ContestsHandler) 

	//scoreboard
	http.HandleFunc("/scoreboard", handlers.ScoreboardHandler)

	//admin
	http.HandleFunc("/admin/submissions",handlers.AdminGetSubmissionsHandler)
	http.HandleFunc("/admin/submissions/{id}", handlers.AdminGetSubmissionHandler)
	http.HandleFunc("/admin/submissions/{id}/rejudge", handlers.AdminRejudgeSubmissionHandler)
	http.HandleFunc("/admin/teams/{id}", handlers.AdminChangeTeamStatusHandler)
	http.HandleFunc("/admin/teams",handlers.AdminGetTeamsHandler)
	http.HandleFunc("/admin/scoreboard",handlers.AdminGetScoreboard)
	http.HandleFunc("/admin/overrides",handlers.AdminGetOverridesHandler)
	http.HandleFunc("/admin/overrides/{id}",handlers.AdminApproveOverride)


	http.HandleFunc("/admin/contest",handlers.AdminChangeContestState)

	ws.StartTimeSyncTicker()

	log.Println("Server listening on :8080")
	err := http.ListenAndServe(":8080", handlers.LoggingMiddleware(cors(http.DefaultServeMux)))
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
