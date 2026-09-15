package handlers

import (
	"campuscode/models"
	"campuscode/service"
	"campuscode/utils"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request body bitch", http.StatusBadRequest)
		return
	}
	//apparently. guys we need such stuff to be sent from frontend so guess its future me problem
	deviceID := r.Header.Get("X-Device-ID") //gpt made this bs i was using uui.new stuff from google
	if deviceID == "" {
		http.Error(w, "missing device id ", http.StatusBadRequest)
		return
	}
	token, err := h.auth.Login(req, deviceID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, "Invalid credentials ", http.StatusUnauthorized)
		case errors.Is(err, service.ErrDisqualified):
			http.Error(w, "team is disqualified", http.StatusForbidden)
		case errors.Is(err, service.ErrDeviceMismatch):
			http.Error(w, "Device mismatch. Request Override validated by Admin", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, 200, map[string]string{
		"session_token": token,
	})
}
