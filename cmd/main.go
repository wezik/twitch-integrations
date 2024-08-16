package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"

	"com.yapdap/pkg/database"
	"com.yapdap/pkg/twitch"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting twitch integration service")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
	} else {
		log.Println("Loaded environment")
	}

	db, err := database.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Println(err)
		} else {
			log.Println("Database closed")
		}
	}(db)

	twitchConn, err := twitch.GetTwitchConnection(db)
	if err != nil {
		log.Fatal(err)
	}
	defer twitchConn.Disconnect()

	twitchConn.SubscribeToCustomRewards()

	<-quit
	log.Println("Shutting down...")
}
