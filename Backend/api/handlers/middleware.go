package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"context"
	"log"
	"net/http"
	"strings"
)

func LoggingMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		log.Printf(" --> %s %s",r.Method,r.URL.Path)
		next.ServeHTTP(w,r)
	})
}

type contextKey string 
const teamContextKey contextKey = "team"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		//check req and give suitable response
		token := r.Header.Get("Authorization")
		token  = strings.TrimPrefix(token,"Bearer ")
		if token == "" {
			http.Error(w,"missing",http.StatusUnauthorized)
			return 
		}
		var session models.Sessions
		err := db.DB.Where("session_token = ? AND active = ?",token,true).First(&session).Error
		if err != nil {
			http.Error(w,"unauthorized",http.StatusUnauthorized)
			return
		}
		var team models.Teams
		//add team into the context so handlers can use it
		db.DB.First(&team,session.TeamID)
		ctx := context.WithValue(r.Context(),teamContextKey,team)
		//if good call next
		next(w,r.WithContext(ctx))
	}
}

func TeamFromContext(r *http.Request) models.Teams {
	team, _ := r.Context().Value(teamContextKey).(models.Teams)
	return team
}