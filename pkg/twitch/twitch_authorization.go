package twitch

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"com.yapdap/pkg/database"
)

const oAuthURL = "https://id.twitch.tv/oauth2/authorize"

var oAuthScopes = []string{
	"user:read:chat",
	"user:write:chat",
	"channel:bot",
	"channel:read:redemptions",
	"channel:manage:redemptions",
}

type OAuthURLRequest struct {
	ClientID string
	RedirectURI string
}

func createOAuthURL(r OAuthURLRequest) string {
	params := url.Values{}
	params.Add("client_id", r.ClientID)
	params.Add("redirect_uri", r.RedirectURI)
	params.Add("response_type", "code")
	scope := strings.Join(oAuthScopes, "+")

	return fmt.Sprintf("%s?%s", oAuthURL, params.Encode()) + "&scope=" + scope
}

type AppTokenRequest struct {
	ClientID string
	ClientSecret string
	DB *sql.DB
}

func GetAppToken(r AppTokenRequest) (string, error) {
	if r.DB == nil {
		panic("DB is nil")
	}

	appToken, err := database.GetToken(r.DB, "app")
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	} else if err == nil && appToken != "" {
		// TODO: Check if the token is valid
		log.Println("Using cached app token")
		return appToken, nil
	}

	authRequest := AuthAppRequest{
		ClientID: r.ClientID,
		ClientSecret: r.ClientSecret,
	}

	appToken, err = authApp(authRequest)
	if err != nil {
		panic(err)
	}

	if err := database.SetToken(r.DB, "app", appToken); err != nil {
		panic(err)
	}

	log.Println("App authorization renewed")

	return appToken, nil
}

type UserTokensRequest struct {
	ClientID string
	ClientSecret string
	RedirectURI string
	DB *sql.DB
}

func GetUserTokens(r UserTokensRequest) (string, string, error) {
	if r.DB == nil {
		panic("DB is nil")
	}

	refreshToken, err := database.GetToken(r.DB, "refresh")
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	} else if err == nil && refreshToken != "" {
		log.Println("Using cached refresh token")
		return "", refreshToken, nil
	}

	oAuthURLRequest := OAuthURLRequest{
		ClientID: r.ClientID,
		RedirectURI: r.RedirectURI,
	}

	oAuthURL := createOAuthURL(oAuthURLRequest)

	log.Println("Opening browser to authorize user at " + oAuthURL)
	err = openURL(oAuthURL)
	if err != nil {
		panic(err)
	}

	code, err := waitForCallback(r.RedirectURI)

	authRequest := AuthUserRequest{
		ClientID: r.ClientID,
		ClientSecret: r.ClientSecret,
		RedirectURI: r.RedirectURI,
		Code: code, // TODO: Get the code from the callback
	}

	res, err := authUser(authRequest)
	if err != nil {
		panic(err)
	}

	if err := database.SetToken(r.DB, "refresh", res.RefreshToken); err != nil {
		panic(err)
	}

	log.Println("User authorization renewed")

	return res.AccessToken, res.RefreshToken, nil
}

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func waitForCallback(redirectURI string) (string, error) {
	uri := strings.Join(strings.Split(redirectURI, ":")[2:], "")
	port := strings.Split(uri, "/")[0]
	mapping := strings.Join(strings.Split(uri, "/")[1:], "/")

	auth := make(chan map[string]string)
	var once sync.Once

	http.HandleFunc("/" + mapping, func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			params := make(map[string]string)
			for key, values := range r.URL.Query() {
				params[key] = values[0]
			}
			auth <- params
			close(auth)
			log.Println("Callback received")
		})
	})

	server := &http.Server{
		Addr: ":" + port,
		ReadTimeout: time.Second * 30,
		WriteTimeout: time.Second * 30,
	}
	
	go func() {
		log.Println("Listening on :" + port + "/" + mapping)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	params := <-auth

	if params == nil {
		return "", errors.New("Callback timed out")
	}

	if err := server.Close(); err != nil {
		return "", err
	}

	code, ok := params["code"]
	if !ok {
		return "", errors.New("No code received")
	}

	return code, nil
}
