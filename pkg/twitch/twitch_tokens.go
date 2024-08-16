package twitch

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"com.yapdap/pkg/database"
	"github.com/nicklaw5/helix/v2"
)

func fetchNewAppToken(client *helix.Client) (string, error) {
	resp, err := client.RequestAppAccessToken([]string{})
	if err != nil {
		return "", err
	}
	return resp.Data.AccessToken, nil
}

func fetchNewUserAndRefreshToken(client *helix.Client) (string, string, error) {
	// temporary solution to be able to run the bot locally
	// TODO: implement redirect to Oauth flow and callback
	userToken, ok := os.LookupEnv("TWITCH_USER_ACCESS_TOKEN")
	if !ok {
		return "", "", fmt.Errorf("TWITCH_USER_ACCESS_TOKEN is not set")
	}

	refreshToken, ok := os.LookupEnv("TWITCH_REFRESH_TOKEN")
	if !ok {
		return "", "", fmt.Errorf("TWITCH_REFRESH_TOKEN is not set")
	}

	return userToken, refreshToken, nil
}

func getAppToken(client *helix.Client, db *sql.DB) (string, error) {
	const APP_TOKEN_ID = "app"
	accessToken, err := database.GetToken(db, APP_TOKEN_ID)

	if err == nil && accessToken != "" {
		// Commented out since it's implemented only for user tokens need custom impl
		// valid, _, err := client.ValidateToken(accessToken)
		// if err != nil && valid {
		// 	log.Println("Using stored app token")
		// 	return accessToken, nil
		// }
		// log.Println("App token is invalid, fetching new one")
		log.Println("Using stored app token")
		return accessToken, nil
	}
	log.Println("App token is invalid, fetching new one")

	accessToken, err = fetchNewAppToken(client)
	if err != nil {
		return "", err
	}
	log.Println("Fetched new app token")

	err = database.SetToken(db, APP_TOKEN_ID, accessToken)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func getRefreshToken(client *helix.Client, db *sql.DB) (string, error) {
	const REFRESH_TOKEN_ID = "refresh"
	refreshToken, err := database.GetToken(db, REFRESH_TOKEN_ID)

	if err == nil && refreshToken != "" {
		// Commented out since it's implemented only for user tokens need custom impl
		// valid, _, err := client.ValidateToken(refreshToken)
		// if err != nil && valid {
		// 	log.Println("Using stored refresh token")
		// 	return refreshToken, nil
		// }
		// log.Println("Refresh token is invalid, fetching new one")
		log.Println("Using stored refresh token")
		return refreshToken, nil
	}
	log.Println("Refresh token is invalid, fetching new one")

	userToken, refreshToken, err := fetchNewUserAndRefreshToken(client)
	if err != nil {
		return "", err
	}
	log.Println("Fetched new user and refresh token")

	err = database.SetToken(db, REFRESH_TOKEN_ID, refreshToken)
	if err != nil {
		return "", err
	}

	client.SetUserAccessToken(userToken)
	return refreshToken, nil
}
