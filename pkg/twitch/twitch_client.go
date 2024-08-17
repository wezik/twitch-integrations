package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	TWITCH_TOKEN_URL = "https://id.twitch.tv/oauth2/token"
)

type AuthAppRequest struct {
	ClientID string
	ClientSecret string
}

func authApp(r AuthAppRequest) (string, error) {
	params := url.Values{}
	params.Add("client_id", r.ClientID)
	params.Add("client_secret", r.ClientSecret)
	params.Add("grant_type", "client_credentials")

	fullURL := fmt.Sprintf("%s?%s", TWITCH_TOKEN_URL, params.Encode())

	resp, err := http.Post(fullURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	appToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Unexpected response from Twitch: %v", result)
	}

	return appToken, nil
}

type AuthUserRequest struct {
	ClientID string
	ClientSecret string
	RedirectURI string
	Code string
}

type AuthUserResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope []string `json:"scope"`
}

func authUser(r AuthUserRequest) (AuthUserResponse, error) {
	params := url.Values{}
	params.Add("client_id", r.ClientID)
	params.Add("client_secret", r.ClientSecret)
	params.Add("grant_type", "authorization_code")
	params.Add("redirect_uri", r.RedirectURI)
	params.Add("code", r.Code)

	fullURL := fmt.Sprintf("%s?%s", TWITCH_TOKEN_URL, params.Encode())

	var response AuthUserResponse

	resp, err := http.Post(fullURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}
