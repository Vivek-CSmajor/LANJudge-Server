package ws

import (
	"sync"
	"time"

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
func broadCast(pl any){
	Mu.RLock()
	defer Mu.RUnlock()
	for conn := range Clients {
		conn.WriteJSON(pl)
	}
}

func BroadcastScoreboard() {
	var scoreboard []models.ScoreBoard
	db.DB.Order("total_solved DESC").Order("total_penalty ASC").Order("last_ac_time ASC").Find(&scoreboard)//seems reduntant or the handler one is reduntant, will hv to check it out
	broadCast(scoreboardMsg{Type: "scoreboard",Rows: scoreboard})
}

func BroadcastTime(){
	if models.ActiveContest == nil {
		return
	}
	broadCast(contestMsg{
		Type: "time_sync",
		Status: models.ActiveContest.Status,
		CurrentTime: time.Now(),
		EndTime: models.ActiveContest.EndTime,
	})
}

func StartTimeSyncTicker() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			BroadcastTime()
		}
	}()
}

type contestMsg struct {
	Type string `json:"type"`
	Status models.ContestStatus `json:"status"`
	CurrentTime time.Time `json:"current_time"`
	EndTime time.Time `json:"end_time"`
}
type scoreboardMsg struct {
	Type string `json:"type"`
	Rows []models.ScoreBoard `json:"rows"`
}



