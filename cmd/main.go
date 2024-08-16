package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"com.yapdap/pkg/twitch"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix/v2"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	twitchConnUser, twitchConnApp, err := twitch.GetTwitchConnection()
	if err != nil {
		panic(err)
	}

	//----

	params := helix.GetCustomRewardsParams{BroadcasterID: twitchConnUser.ChannelID}
	response, err := twitchConnUser.Client.GetCustomRewards(&params)

	if err != nil {
		log.Println(err)
	}
	for _, emote := range response.Data.ChannelCustomRewards {
		log.Printf("Custom reward: %s/%s", emote.ID, emote.Title)
	}

	//----

	resp, err := twitchConnApp.Client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type: helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: twitchConnApp.ChannelID,
			UserID: twitchConnApp.ChannelID,
		},
		Transport: helix.EventSubTransport{
			Method: "webhook",
			Callback: "https://localhost:443/eventsub",
			Secret: "1234567890",
		},

	})
	if err != nil {
		log.Println(err)
	}

	log.Printf("%+v\n", resp)


	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	// read input and pass to chat
	// 	message, err := reader.ReadString('\n')
	// 	message = strings.TrimSuffix(message, "\n")
	//
	// 	if err != nil {
	// 		log.Println("Error reading message")
	// 		continue
	// 	}
	//
	// 	response := sendMessageToChat(twitchConn, message)
	// 	if response.StatusCode != 200 {
	// 		log.Println("Error sending message")
	// 	} else {
	// 		log.Printf("Message sent, rate limit: %v/%v\n", response.GetRateLimitRemaining(), response.GetRateLimit())
	// 	}
	// }

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println(" Shutting down...")
}

type eventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Event        json.RawMessage            `json:"event"`
	Challenge    string                     `json:"challenge"`
}

func eventSubMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()

	if !helix.VerifyEventSubNotification("1234567890", r.Header, string(body)) {
		log.Println("No valid signature on subscription")
		return
	} else {
		log.Println("Valid signature on subscription")
	}

	var vals eventSubNotification
	err = json.Unmarshal(body, &vals)
	if err != nil {
		log.Println(err)
		return
	}

	if vals.Challenge != "" {
		w.Write([]byte(vals.Challenge))
		return
	}

	var messageEvent helix.EventSubChatMessage
	err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&messageEvent)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Received message from %s: %v", messageEvent.Text, messageEvent.Fragments)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func sendMessageToChat(tConn *twitch.TwitchConn, message string) helix.ResponseCommon {
	params := helix.SendChatMessageParams{
		BroadcasterID: tConn.ChannelID,
		SenderID:      tConn.ChannelID,
		Message:       message,
	}

	log.Printf("Sending chat message to %s room: %s\n", params.BroadcasterID, params.Message)
	response, err := tConn.Client.SendChatMessage(&params)
	if err != nil {
		log.Println(err)
	}
	return response.ResponseCommon
}
