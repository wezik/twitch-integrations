package twitch

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"com.yapdap/pkg/database"
	"github.com/nicklaw5/helix/v2"
)

type TwitchConn struct {
	UserClient  *helix.Client
	AppClient   *helix.Client
	Broadcaster *helix.User
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
	val, ok = os.LookupEnv("TWITCH_REFRESH_TOKEN")
	if !ok || val == "" {
		return nil, fmt.Errorf("TWITCH_REFRESH_TOKEN not set")
	}
	helixOptions.RefreshToken = val
	return &helixOptions, nil
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

	helixOptionsUser, err := getHelixOptionsUser()
	if err != nil {
		return nil, err
	}

	helixOptionsApp, err := getHelixOptions()
	if err != nil {
		return nil, err
	}

	log.Println("Helix options fetched")

	userClient, err := createUserClient(helixOptionsUser)
	if err != nil {
		return nil, err
	}
	log.Println("Fetched new user access token")

	userClient.OnUserAccessTokenRefreshed(func(_, _ string) {
		log.Println("")
	})

	appClient, err := createAppClient(helixOptionsApp)
	if err != nil {
		return nil, err
	}
	log.Println("App client created")

	accessToken, err := getAppAccessToken(appClient, db)
	if err != nil {
		return nil, err
	}
	appClient.SetAppAccessToken(accessToken)

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

var channelsCache = map[string]*helix.User{}

func createUserClient(helixOptions *helix.Options) (*helix.Client, error) {
	client, err := helix.NewClient(helixOptions)
	if err != nil {
		return nil, err
	}
	client.OnUserAccessTokenRefreshed(func(_, _ string) {
		log.Println("Token refreshed")
	})
	return client, nil
}

func createAppClient(helixOptions *helix.Options) (*helix.Client, error) {
	client, err := helix.NewClient(helixOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getBroadcaster(client *helix.Client, login string) (*helix.User, error) {
	if user, ok := channelsCache[login]; ok {
		return user, nil
	}
	params := helix.UsersParams{Logins: []string{login}}
	channels, err := client.GetUsers(&params)
	if err != nil {
		return nil, err
	}

	for _, user := range channels.Data.Users {
		if user.Login == login {
			channelsCache[login] = &user
			return &user, nil
		}
	}
	return nil, fmt.Errorf("User not found")
}

func getAppAccessToken(client *helix.Client, db *sql.DB) (string, error) {
	accessToken, err := database.GetToken(db, "app")
	if err == nil && accessToken != "" {
		log.Println("Using stored app access token")
		return accessToken, nil
	}

	accessToken, err = fetchNewAccessToken(client)
	if err != nil {
		return "", err
	}
	err = database.SetToken(db, "app", accessToken)
	if err != nil {
		return "", err
	}
	log.Println("Fetched new app access token")
	return accessToken, nil

}
