package twitch

import (
	"github.com/nicklaw5/helix/v2"
)

func fetchNewAccessToken(client *helix.Client) (string, error) {
	resp, err := client.RequestAppAccessToken([]string{})
	if err != nil {
		return "", err
	}
	return resp.Data.AccessToken, nil
}
