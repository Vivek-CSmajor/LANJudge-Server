package handlers

import (
	"campuscode/api/ws"
	"campuscode/db"
	"campuscode/utils"
	"encoding/json"

	"log"

	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ScoreboardHandler(w http.ResponseWriter, r *http.Request) {
	//a fall back to allow plain http fetches instead of a ws always
	if !websocket.IsWebSocketUpgrade(r) {
		scoreboard, err := db.GetScoreboard()
		if err != nil {
			http.Error(w, "Db error when sorting", http.StatusInternalServerError)
			return
		}
		utils.WriteJSON(w, 200, scoreboard)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error in websocket ", err)
		return
	}
	ws.Mu.Lock()
	ws.Clients[conn] = true //ow my backend will remember this connection and if it exists
	ws.Mu.Unlock()
	defer conn.Close()
	scoreboard, err := db.GetScoreboard()
	if err != nil {
		http.Error(w, "Db error when sorting", http.StatusInternalServerError)
		return
	}
	log.Println(scoreboard)
	//the conn writejson below could be done in a forloop to broadcast the scorebaord to every active connection
	// so we wouldnt have to write the lower conn.WriteJson(scorebaord) thing and just mark the conn = true
	b, _ := json.MarshalIndent(scoreboard, "", " ")
	log.Println("Scoreboard payload : \n", string(b))
	err = conn.WriteJSON(scoreboard)
	if err != nil {
		http.Error(w, "WS error when writing json scoreboard to conn", http.StatusInternalServerError)
		return
	}
	for {
		_, _, err := conn.ReadMessage() //sends close connection message as error when browser disconnectes so to break use the error below
		if err != nil {
			ws.Mu.Lock()
			delete(ws.Clients, conn)
			ws.Mu.Unlock()
			break
		}
	}
	//utils.WriteJSON(w, 200, scoreboard)
}
