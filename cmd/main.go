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
	log.Println("Starting yapdap twitch-obs integration")
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
	} else {
		log.Println("Loaded .env")
	}

	db, err := database.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println(err)
		} else {
			log.Println("Database closed")
		}
	}()

	twitchConn, err := twitch.GetTwitchConnection(db)
	if err != nil {
		log.Fatal(err)
	}
	defer twitchConn.Disconnect()

	twitchConn.SubscribeToCustomRewards()

	<-quit
}

func onQuit() {

}
