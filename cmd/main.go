package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"com.yapdap/pkg/twitch"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting...")
	log.Println("Loading .env file without overriding existing variables...")
	if err := godotenv.Load(".env"); err != nil {
		log.Println(err)
	}

	twitchConn, err := twitch.GetTwitchConnection()
	if err != nil {
		log.Fatal(err)
	}

	twitchConn.SubscribeToCustomRewards()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println()
	log.Println("Shutting down...")
	log.Println("Invalidating user access token...")
	token := twitchConn.UserClient.GetUserAccessToken()
	twitchConn.UserClient.RevokeUserAccessToken(token)
}
