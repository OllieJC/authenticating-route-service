package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func OAuthGoogleLogin(r *http.Response) {

	// Create oauthState cookie
	oauthState := GenerateStateOauthCookie(r)

	/*
	   AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
	   validate that it matches the the state query parameter on your redirect callback.
	*/
	u := googleOauthConfig.AuthCodeURL(oauthState)

	RedirectResponse(r, http.StatusTemporaryRedirect, u)
}

func OauthGoogleCallback(r *http.Request, resp *http.Response) {
	// Read oauthState from Cookie
	oauthState, err := r.Cookie("oauthstate")

	if err != nil || r.FormValue("state") != oauthState.Value {
		RedirectResponse(resp, http.StatusTemporaryRedirect, "/auth/login")
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		RedirectResponse(resp, http.StatusTemporaryRedirect, "/auth/login")
	}

	if len(data) < 0 {
		// GetOrCreate User in your db.
		// Redirect or response with a token.
		// More code .....

		RedirectResponse(resp, http.StatusTemporaryRedirect, "/")
		// set cookie
		fmt.Println("data:", data)
	}
}

func GenerateStateOauthCookie(resp *http.Response) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b, err := generateRandomBytes(16)
	if err != nil {
		panic(fmt.Sprintf("generateRandomBytes is unavailable: failed with %#v", err))
	}

	state := base64.StdEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}

	resp.Header.Add("Set-Cookie", cookie.String())

	return state
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}
