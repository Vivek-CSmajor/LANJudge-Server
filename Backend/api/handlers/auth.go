package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"time"

	//"campuscode/utils"
	"encoding/json"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	//3 cases
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method called in login Handler", http.StatusMethodNotAllowed)
		return
	}
	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request body bitch", http.StatusBadRequest)
		return
	}
	//this already checks the auth part
	team, err := db.GetTeamByNameAndLoginCode(req.Name, req.LoginCode)
	if err != nil {
		http.Error(w, "Invalid credentials ", http.StatusUnauthorized)
		return
	}
	if team.Status != nil && *team.Status == "disqualified" {
		http.Error(w, "team is disqualified", http.StatusForbidden)
		return
	}
	//apparently. guys we need such stuff to be sent from frontend so guess its future me problem
	deviceID := r.Header.Get("X-Device-ID") //gpt made this bs i was using uui.new stuff from google
	if deviceID == "" {
		http.Error(w, "missing device id ", 400)
		return
	}
	login := func() {
		if err := db.DeactivateTeamSessions(team.ID); err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		sessionToken := utils.TokenGenerator(16)
		session := models.Sessions{
			TeamID:       team.ID,
			SessionToken: sessionToken,
			DeviceID:     deviceID,
			Active:       true,
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(6 * time.Hour), /// currTime + 6hrs
		}
		if err := db.CreateSession(session); err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		utils.WriteJSON(w, 200, map[string]string{
			"session_token": sessionToken,
		})
	}
	//Case 1 : First time logging in so no existing session
	if team.BoundDeviceID == nil {
		log.Println("Babes first time log in ")
		if err := db.UpdateTeamBoundDevice(team.ID, deviceID); err != nil {
			http.Error(w, "Failed to bind device", http.StatusInternalServerError)
			return
		}
		login()
		return

	} else if team.BoundDeviceID != nil && *team.BoundDeviceID != deviceID {
		//Case2 : Login again but with diff device
		//call override
		if err := db.CreateDeviceOverride(models.DeviceOverride{TeamID: team.ID, NewDeviceID: deviceID, Status: "pending", CreatedAt: time.Now()}); err != nil {
			http.Error(w, "Failed to request override", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Device mismatch. Request Override validated by Admin", http.StatusUnauthorized)
		return

	} else if team.BoundDeviceID != nil && *team.BoundDeviceID == deviceID {
		login()
		return
	}
}
