package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	// "os/signal"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix/v2"
)

type State struct {
        Client *helix.Client
        ChannelID string
}

var savedChannels = map[string]string{
        "brzdyngol": "31809634",
}

var state State

func loadOptions() helix.Options{
        var helixOptions helix.Options 
        val, ok := os.LookupEnv("TWITCH_CLIENT_ID")
        if !ok || val == "" {
                panic("TWITCH_CLIENT_ID not set")
        }
        helixOptions.ClientID = val
        val, ok = os.LookupEnv("TWITCH_CLIENT_SECRET")
        if !ok || val == "" {
                panic("TWITCH_CLIENT_SECRET not set")
        }
        helixOptions.ClientSecret = val
        val, ok = os.LookupEnv("TWITCH_USER_ACCESS_TOKEN")
        if !ok || val == "" {
                panic("TWITCH_USER_ACCESS_TOKEN not set")
        }
        helixOptions.UserAccessToken = val
        val, ok = os.LookupEnv("TWITCH_REFRESH_TOKEN")
        if !ok || val == "" {
                panic("TWITCH_REFRESH_TOKEN not set")
        }
        helixOptions.RefreshToken = val
        return helixOptions
}

func main() {
        if err := godotenv.Load(".env"); err != nil {
                panic(err)
        }
        helixOptions := loadOptions()

        userLogin, ok := os.LookupEnv("TWITCH_USER_NAME")
        if !ok || userLogin == "" {
                panic("TWITCH_USER_NAME not set")
        }

        client, err := helix.NewClient(&helixOptions)
        if err != nil {
                panic(err)
        } else {
                fmt.Println("Client created")
        }

        channelID, err := getChannelID(client, userLogin)
        if err != nil {
                panic(err)
        } else {
                fmt.Println("Channel ID for " + userLogin + " is " + channelID)
        }

        state = State{Client: client, ChannelID: channelID}
        fmt.Println("State created")

        reader := bufio.NewReader(os.Stdin)
        for {
                // read input and pass to chat
                message, err := reader.ReadString('\n')
                message = strings.TrimSuffix(message, "\n")

                if err != nil {
                        fmt.Println("Error reading message")
                        continue
                }

                response := sendMessageToChat(message)
                if response.StatusCode != 200 {
                        fmt.Println("Error sending message")
                } else {
                        fmt.Printf("Message sent, rate limit: %v/%v\n", response.GetRateLimitRemaining(), response.GetRateLimit())
                }

        }

        // quit := make(chan os.Signal, 1)
        // signal.Notify(quit, os.Interrupt)
        // <-quit
        //
        // fmt.Println(" Shutting down...")
}

func registerWebhook() {
        r := mux.NewRouter()

        r.HandleFunc("/eventsub", webhookHandler).Methods("POST")

        srv := &http.Server{
                Handler: r,
                Addr:    ":443",
        }

        fmt.Printf("Starting webhook server on port %v\n", 443)
        err := srv.ListenAndServe()
        if err != nil {
                panic(err)
        }
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        err := r.ParseForm()
        if err != nil {
                http.Error(w, "Error parsing form", http.StatusBadRequest)
                return
        }

        fmt.Printf("Received webhook: %s\n", r.Form.Get("data"))
}

func sendMessageToChat(message string) helix.ResponseCommon {
        params := helix.SendChatMessageParams{
                BroadcasterID: state.ChannelID,
                SenderID:      state.ChannelID,
                Message: message,
        }

        fmt.Printf("Sending chat message to %s room: %s\n", params.BroadcasterID, params.Message)
        response, err := state.Client.SendChatMessage(&params)
        if err != nil {
                fmt.Println(err)
        }
        return response.ResponseCommon
}

func getChannelID(client *helix.Client, channelName string) (string, error) {
        if channelID, ok := savedChannels[channelName]; ok {
                return channelID, nil
        }
        params := helix.SearchChannelsParams{Channel: channelName}
        channels, err := client.SearchChannels(&params)
        if err != nil {
                return "", err
        }

        if len(channels.Data.Channels) == 0 {
                return "", fmt.Errorf("No channels found")
        }

        for _, channel := range channels.Data.Channels {
                if channel.BroadcasterLogin == channelName {
                        savedChannels[channelName] = channel.ID
                        return channel.ID, nil
                }
        }
        return "", fmt.Errorf("Channel not found")
}
