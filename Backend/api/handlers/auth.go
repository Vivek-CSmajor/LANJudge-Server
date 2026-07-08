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

func LoginHandler(w http.ResponseWriter, r *http.Request){
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
	var team models.Teams

	//this already checks the auth part
	err = db.DB.Where("name = ? AND login_code = ?",req.Name,req.LoginCode).First(&team).Error
	if err != nil {
		http.Error(w, "Invalid credentials ", http.StatusUnauthorized)
		return
	}
	//apparently. guys we need such stuff to be sent from frontend so guess its future me problem
	deviceID := r.Header.Get("X-Device-ID")//gpt made this bs i was using uui.new stuff from google
	if deviceID == ""{
		http.Error(w, "mising device id ",400)
	}
	login := func(){
		db.DB.Model(&models.Sessions{}).
		Where("team_id = ?",team.ID).
		Update("active",false)
		
		sessionToken := utils.TokenGenerator(16)
		session := models.Sessions{
			TeamID: team.ID,
			SessionToken: sessionToken,
			DeviceID: deviceID,
			Active: true,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(6*time.Hour),/// currTime + 6hrs			
		}
		db.DB.Create(&session)
		utils.WriteJSON(w, 200, map[string]string{
			"session_token" : sessionToken,
		})
	}
	//Case 1 : First time logging in so no existing session
	if(team.BoundDeviceID == nil){
		log.Println("Babes first time log in ")
		team.BoundDeviceID = &deviceID
		db.DB.Save(&team)
		login()
		return

	}else if(team.BoundDeviceID != nil && *team.BoundDeviceID != deviceID){
		//Case2 : Login again but with diff device
		//call override
		http.Error(w,"Device mismatch. Request Override validated by Admin",http.StatusUnauthorized)
		return
		
	}else if(team.BoundDeviceID != nil && *team.BoundDeviceID == deviceID){
		login()
		return 
	}
}