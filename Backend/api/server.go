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

func Server(h *handlers.Handlers) {
	http.HandleFunc("GET /health", handlers.HealthHandler)

	http.HandleFunc("POST /submit", handlers.RequireAuth(handlers.RequireContestOpen(h.SubmitHandler)))
	http.HandleFunc("GET /submission/{id}", handlers.RequireAuth(h.GetSubmissionHandler))
	http.HandleFunc("GET /submissions/me", handlers.RequireAuth(h.ListMySubmissionsHandler))
	http.HandleFunc("POST /run", handlers.RequireAuth(h.RunHandler))

	//auth rotues
	http.HandleFunc("POST /register", h.RegisterHandler)
	http.HandleFunc("POST /login", h.LoginHandler)
	//http.HandleFunc("/logout", handlers.LogoutHandler)

	http.HandleFunc("GET /problems", handlers.RequireAuth(handlers.RequireContestOpen(h.ProblemsHandler)))
	http.HandleFunc("GET /problems/{id}", handlers.RequireAuth(handlers.RequireContestOpen(h.ProblemHandler)))

	//contest handlers
	http.HandleFunc("GET /contest", handlers.RequireAuth(h.ContestHandler))

	//scoreboard
	http.HandleFunc("GET /scoreboard", h.ScoreboardHandler)

	//admin
	http.HandleFunc("POST /admin/login", h.AdminLoginHandler)
	http.HandleFunc("GET /admin/submissions", handlers.RequireAdminAuth(h.AdminGetSubmissionsHandler))
	http.HandleFunc("GET /admin/submissions/{id}", handlers.RequireAdminAuth(h.AdminGetSubmissionHandler))
	http.HandleFunc("POST /admin/submissions/{id}/rejudge", handlers.RequireAdminAuth(h.AdminRejudgeSubmissionHandler))
	http.HandleFunc("PATCH /admin/teams/{id}", handlers.RequireAdminAuth(h.AdminChangeTeamStatusHandler))
	http.HandleFunc("GET /admin/teams", handlers.RequireAdminAuth(h.AdminGetTeamsHandler))
	http.HandleFunc("GET /admin/scoreboard", handlers.RequireAdminAuth(h.AdminGetScoreboard))
	http.HandleFunc("GET /admin/overrides", handlers.RequireAdminAuth(h.AdminGetOverridesHandler))
	http.HandleFunc("POST /admin/overrides/{id}", handlers.RequireAdminAuth(h.AdminApproveOverride))
	http.HandleFunc("GET /admin/contest", handlers.RequireAdminAuth(h.AdminGetContestStateHandler))
	http.HandleFunc("PATCH /admin/contest", handlers.RequireAdminAuth(h.AdminChangeContestState))

	ws.StartTimeSyncTicker()

	log.Println("Server listening on :8080")
	err := http.ListenAndServe(":8080", handlers.LoggingMiddleware(cors(http.DefaultServeMux)))
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
