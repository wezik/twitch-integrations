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

func GetTwitchConnection(db *sql.DB) (*TwitchConn, error) {
	log.Println("Establishing Twitch connection...")

	log.Println("Fetching helix options...")
	helixOptionsUser, err := getHelixOptionsUser()
	if err != nil {
		return nil, err
	}

	helixOptionsApp, err := getHelixOptions()
	if err != nil {
		return nil, err
	}
	log.Println("Helix options fetched")

	log.Println("Creating clients...")

	userClient, err := createUserClient(helixOptionsUser)
	if err != nil {
		return nil, err
	}
	log.Println("User client created")

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

	log.Println("Fetching broadcaster...")
	userLogin, ok := os.LookupEnv("TWITCH_USER_NAME")
	if !ok || userLogin == "" {
		return nil, fmt.Errorf("TWITCH_USER_NAME not set")
	}
	broadcaster, err := getBroadcaster(appClient, userLogin)
	if err != nil {
		return nil, err
	}
	log.Printf("User fetched [%s, %s, %s]\n", broadcaster.Login, broadcaster.DisplayName, broadcaster.ID)

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
	log.Println("Checking stored access token...")
	accessToken, err := database.GetToken(db, "app")
	if err == nil && accessToken != "" {
		log.Println("Access token found")
		return accessToken, nil
	}

	log.Println("Fetching access token...")
	accessToken, err = fetchNewAccessToken(client)
	if err != nil {
		return "", err
	}
	err = database.SetToken(db, "app", accessToken)
	if err != nil {
		return "", err
	}
	log.Println("Access token fetched")
	return accessToken, nil

}
