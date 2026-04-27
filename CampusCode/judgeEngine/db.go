package judgeEngine

import (

	"log"

	//"github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(){
	db, err := gorm.Open(sqlite.Open("server_data/runtime/db.sqlite"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Println("DB connection established")
	db.AutoMigrate(&Submission{})
	
	DB = db
}