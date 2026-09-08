package handlers

import (
	"campuscode/utils"
	"net/http"
	"time"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Wrong method is called", http.StatusMethodNotAllowed)
		return
	}
	utils.WriteJSON(w, 200, map[string]any{
		"status": "ok",
		"time":   time.Now(),
	})
}
