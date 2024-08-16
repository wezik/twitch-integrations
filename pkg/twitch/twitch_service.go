package twitch

import (
	"fmt"
	"log"
	"os"

	"github.com/nicklaw5/helix/v2"
)

type TwitchConn struct {
	Client    *helix.Client
	ChannelID string
}

func getHelixOptionsUser() (*helix.Options, error) {
	var helixOptions helix.Options
	val, ok := os.LookupEnv("TWITCH_CLIENT_ID")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID not set")
	}
	helixOptions.ClientID = val
	val, ok = os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET not set")
	}
	helixOptions.ClientSecret = val
	val, ok = os.LookupEnv("TWITCH_USER_ACCESS_TOKEN")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_USER_ACCESS_TOKEN not set")
	}
	helixOptions.UserAccessToken = val
	val, ok = os.LookupEnv("TWITCH_REFRESH_TOKEN")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_REFRESH_TOKEN not set")
	}
	helixOptions.RefreshToken = val
	return &helixOptions, nil
}

func getHelixOptionsApp() (*helix.Options, error) {
	var helixOptions helix.Options
	val, ok := os.LookupEnv("TWITCH_CLIENT_ID")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID not set")
	}
	helixOptions.ClientID = val
	val, ok = os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET not set")
	}
	helixOptions.ClientSecret = val
	val, ok = os.LookupEnv("TWITCH_APP_ACCESS_TOKEN")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_APP_ACCESS_TOKEN not set")
	}
	helixOptions.AppAccessToken = val
	return &helixOptions, nil
}

func GetTwitchConnection() (*TwitchConn, *TwitchConn, error) {
	userLogin, ok := os.LookupEnv("TWITCH_USER_NAME")
	if !ok || userLogin == "" {
		return nil, nil, fmt.Errorf("TWITCH_USER_NAME not set")
	}
	helixOptionsUser, err := getHelixOptionsUser()
	if err != nil {
		return nil, nil, err
	}
	clientUser, err := helix.NewClient(helixOptionsUser)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Client created")
	clientUser.OnUserAccessTokenRefreshed(func(_, _ string) {
		log.Println("Token refreshed")
	})
	channelID, err := GetChannelID(clientUser, userLogin)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("User %s found with channel ID %s\n", userLogin, channelID)
	helixOptionsApp, err := getHelixOptionsApp()
	if err != nil {
		return nil, nil, err
	}
	clientApp, err := helix.NewClient(helixOptionsApp)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Client created")
	log.Println("Twitch connection established")
	return &TwitchConn{Client: clientUser, ChannelID: channelID}, &TwitchConn{Client: clientApp, ChannelID: channelID}, nil
}

var channelsCache = map[string]string{
	"brzdyngol": "31809634",
}

func GetChannelID(client *helix.Client, channelName string) (string, error) {
	if channelID, ok := channelsCache[channelName]; ok {
		return channelID, nil
	}
	params := helix.SearchChannelsParams{Channel: channelName}
	channels, err := client.SearchChannels(&params)
	if err != nil {
		return "", err
	}
	for _, channel := range channels.Data.Channels {
		if channel.BroadcasterLogin == channelName {
			channelsCache[channelName] = channel.ID
			return channel.ID, nil
		}
	}
	return "", fmt.Errorf("Channel not found")
}
