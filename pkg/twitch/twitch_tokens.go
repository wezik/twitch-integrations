package twitch

import (
	"net/http"
)

const (
	OAUTH2_ENDPOINT = "https://id.twitch.tv/oauth2/authorize"
	OAUTH2_REDIRECT_URI = "http://localhost:3030/userauth"
)

const VALIDATE_ENDPOINT = "https://id.twitch.tv/oauth2/validate"

func isTokenValid(token string) (bool, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, VALIDATE_ENDPOINT, nil)
	req.Header.Add("Authorization", "Bearer "  + token)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func revokeToken(clientID string, token string) error {
	return nil
}
