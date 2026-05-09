package handlers

import (
	"campuscode/db"
	"campuscode/models"
	"campuscode/utils"
	"log"

	//"campuscode/utils"
	"net/http"
)

func ScoreboardHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w,"invalid method is called for scoreboard", http.StatusMethodNotAllowed)
		return 
	}
	var scoreboard []models.ScoreBoard
	db.DB.Order("total_solved DESC").Order("total_penalty ASC").
	Order("last_ac_time ASC").Find(&scoreboard)
	log.Println(scoreboard)
	utils.WriteJSON(w, 200, scoreboard)	
}

