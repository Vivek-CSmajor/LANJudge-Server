package db

import (
	"campuscode/models"
	"log"
	//"github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var conn *gorm.DB

func InitDB() {
	db, err := gorm.Open(sqlite.Open("server_data/runtime/db.sqlite"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Println("DB connection established")
	if err := db.AutoMigrate(&models.Submission{}, &models.Teams{}, &models.Sessions{}, &models.DeviceOverride{}, &models.Contest{}, &models.TeamProblemStats{}, &models.ScoreBoard{}, &models.Problem{}); err != nil {
		panic(err)
	}

	conn = db
}
