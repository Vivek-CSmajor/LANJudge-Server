package ws

import (
	"sync"

	"campuscode/db"
	"campuscode/models"

	"github.com/gorilla/websocket"
)

var Clients = map[*websocket.Conn]bool{}
var Mu sync.RWMutex
//a map to store the connections (websockets) 

//above:its better to use RWMutex -->allows more than 1 reads but blocks more than 1 writes at same time. 
// normal sync.Mutex blocks concurrent reads too and since we gonna just read its better to have RWMutex
// ok i i m also writing thge connection so its better to use sync.Mutex but no harm in using RWMutex
func BroadcastScoreboard() {
	var scoreboard []models.ScoreBoard
	db.DB.Order("total_solved DESC").Order("total_penalty ASC").Order("last_ac_time ASC").Find(&scoreboard)
	Mu.RLock()
	defer Mu.RUnlock()
	for conn := range Clients {
		conn.WriteJSON(scoreboard)
	}
}
