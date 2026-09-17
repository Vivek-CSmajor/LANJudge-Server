package handlers

import (
	"campuscode/contest"
	"campuscode/db"
	"campuscode/models"
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(" --> %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const teamContextKey contextKey = "team"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check req and give suitable response
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			http.Error(w, "missing", http.StatusUnauthorized)
			return
		}
		session, err := db.GetActiveSessionByToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		//add team into the context so handlers can use it
		team, _ := db.GetByID[models.Teams](session.TeamID)
		if allowed, reason := team.CanSubmit(); !allowed {
			http.Error(w, reason, http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), teamContextKey, team)
		//if good call next
		next(w, r.WithContext(ctx))
	}
}

func TeamFromContext(r *http.Request) models.Teams {
	team, _ := r.Context().Value(teamContextKey).(models.Teams)
	return team
}

func RequireContestOpen(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !contest.Active.IsOpen(time.Now()) {
			http.Error(w, "contest is currently not open", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check req and give suitable response
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			http.Error(w, "missing", http.StatusUnauthorized)
			return
		}
		_, err := db.GetActiveAdminSessionByToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		//if good call next
		next(w, r)
	}
}
