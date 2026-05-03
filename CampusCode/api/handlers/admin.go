package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"encoding/json"
	"log"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method is called , must call POST", http.StatusMethodNotAllowed)
		return
	}
	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"Invalid Request body ", http.StatusBadRequest)
		return
	}
	loginCode := utils.TokenGenerator(8)
	res := models.RegisterResponse{
		Name : req.Name,
		LoginCode: loginCode,
	}
	db.RegisterTeamRepository(res)
	log.Println("Team Registered")
	utils.WriteJSON(w ,200,res)
}

