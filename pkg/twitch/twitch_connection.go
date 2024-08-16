package twitch

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/nicklaw5/helix/v2"
)

type TwitchConn struct {
	UserClient  *helix.Client
	AppClient   *helix.Client
	Broadcaster *helix.User
}

func (tc *TwitchConn) Disconnect() {
	token := tc.UserClient.GetUserAccessToken()
	_, err := tc.UserClient.RevokeUserAccessToken(token)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Sent revoke user token request")
	}
	log.Println("Disconnected from Twitch")
}

func GetTwitchConnection(db *sql.DB) (*TwitchConn, error) {
	log.Println("Establishing Twitch connection")

	helixOptionsUser, err := getHelixOptions()
	if err != nil {
		return nil, err
	}

	helixOptionsApp := helixOptionsUser

	log.Println("Helix options fetched")

	// User client

	userClient, err := helix.NewClient(helixOptionsUser)
	if err != nil {
		return nil, err
	}
	log.Println("User client created")

	userClient.OnUserAccessTokenRefreshed(func(_, _ string) {
		log.Println("User token renewed")
	})

	refreshToken, err := getRefreshToken(userClient, db)
	if err != nil {
		return nil, err
	}
	userClient.SetRefreshToken(refreshToken)

	// App client

	appClient, err := helix.NewClient(helixOptionsApp)
	if err != nil {
		return nil, err
	}
	log.Println("App client created")

	accessToken, err := getAppToken(appClient, db)
	if err != nil {
		return nil, err
	}
	appClient.SetAppAccessToken(accessToken)

	// Broadcaster fetch

	log.Println("Twitch connection established")

	userLogin, ok := os.LookupEnv("TWITCH_USER_NAME")
	if !ok || userLogin == "" {
		return nil, fmt.Errorf("TWITCH_USER_NAME not set")
	}
	broadcaster, err := getBroadcaster(appClient, userLogin)
	if err != nil {
		return nil, err
	}
	log.Printf("User set to [%s, %s]\n", broadcaster.DisplayName, broadcaster.ID)

	return &TwitchConn{UserClient: userClient, AppClient: appClient, Broadcaster: broadcaster}, nil
}

func getHelixOptions() (*helix.Options, error) {
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
	return &helixOptions, nil
}

func getBroadcaster(client *helix.Client, login string) (*helix.User, error) {
	params := helix.UsersParams{Logins: []string{login}}
	channels, err := client.GetUsers(&params)
	if err != nil {
		return nil, err
	}

	for _, user := range channels.Data.Users {
		if user.Login == login {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("User not found")
}
