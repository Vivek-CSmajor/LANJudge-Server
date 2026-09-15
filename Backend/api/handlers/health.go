package handlers

import (
	"campuscode/utils"
	"net/http"
	"time"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, 200, map[string]any{
		"status": "ok",
		"time":   time.Now(),
	})
}
