package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"com.yapdap/pkg/twitch"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix/v2"
)

func main() {
        if err := godotenv.Load(".env"); err != nil {
                panic(err)
        }

        twitchConn, err := twitch.GetTwitchConnection()
        if err != nil {
                panic(err)
        }

        //----
        
        params := helix.GetChannelEmotesParams{BroadcasterID: twitchConn.ChannelID}
        response, err := twitchConn.Client.GetChannelEmotes(&params)

        if err != nil {
                log.Println(err)
        }
        for _, emote := range response.Data.Emotes {
                log.Printf("Emote: %s/%s, ", emote.ID, emote.Name)
        }

        //----

        reader := bufio.NewReader(os.Stdin)
        for {
                // read input and pass to chat
                message, err := reader.ReadString('\n')
                message = strings.TrimSuffix(message, "\n")

                if err != nil {
                        log.Println("Error reading message")
                        continue
                }

                response := sendMessageToChat(twitchConn, message)
                if response.StatusCode != 200 {
                        log.Println("Error sending message")
                } else {
                        log.Printf("Message sent, rate limit: %v/%v\n", response.GetRateLimitRemaining(), response.GetRateLimit())
                }

        }

        // quit := make(chan os.Signal, 1)
        // signal.Notify(quit, os.Interrupt)
        // <-quit
        //
        // fmt.Println(" Shutting down...")
}

func sendMessageToChat(tConn *twitch.TwitchConn, message string) helix.ResponseCommon {
        params := helix.SendChatMessageParams{
                BroadcasterID: tConn.ChannelID,
                SenderID:      tConn.ChannelID,
                Message: message,
        }

        log.Printf("Sending chat message to %s room: %s\n", params.BroadcasterID, params.Message)
        response, err := tConn.Client.SendChatMessage(&params)
        if err != nil {
                log.Println(err)
        }
        return response.ResponseCommon
}
