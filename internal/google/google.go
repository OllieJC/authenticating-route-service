package google

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	c "authenticating-route-service/internal/configurator"
	h "authenticating-route-service/internal/httphelper"
	u "authenticating-route-service/internal/utils"
	. "authenticating-route-service/pkg/debugprint"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ProviderString is Google
const ProviderString = "google"
const redirectFormatString = "%s://%s/auth/callback/google/%s"

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var googleOauthConfig = &oauth2.Config{
	Scopes:   []string{"profile", "email", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: google.Endpoint,
}

func setOauthConfig(dc c.DomainConfig, emailDomain string) {
	setVal := ""
	if emailDomain != "" {
		if dc.Domain != "" {
			setVal = fmt.Sprintf(redirectFormatString, "https", dc.Domain, emailDomain)
			googleOauthConfig.RedirectURL = setVal
			Debugfln("setOauthConfig: Setting RedirectURL to: %s", setVal)
		}
		gled := dc.GetLoginEmailDomain(emailDomain, ProviderString)
		if gled.Provider == "google" {
			googleOauthConfig.ClientID = gled.OAuthClientID
			googleOauthConfig.ClientSecret = gled.OAuthClientSecret
		}
	}
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func OAuthGoogleLogin(response *http.Response, dc c.DomainConfig, emailDomain string) {
	Debugfln("OAuthGoogleLogin:1: Start...")

	// Create oauthState cookie
	oauthState := GenerateStateOauthCookie(response)

	/*
	   AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
	   validate that it matches the the state query parameter on your redirect callback.
	*/
	setOauthConfig(dc, emailDomain)
	oAuthUrl := googleOauthConfig.AuthCodeURL(oauthState)

	h.RedirectResponse(response, http.StatusSeeOther, oAuthUrl)
}

func OauthGoogleCallback(request *http.Request, response *http.Response, dc c.DomainConfig) (string, error) {
	Debugfln("OauthGoogleCallback:1: Start...")

	oauthState, err := request.Cookie("oauthstate")

	if err != nil {
		Debugfln("OauthGoogleCallback:err: %#v", err)
		return "", fmt.Errorf("ERROR: OauthGoogleCallback: state bad")
	} else if request.FormValue("state") != oauthState.Value {
		Debugfln("OauthGoogleCallback:err: state bad")
		return "", fmt.Errorf("ERROR: OauthGoogleCallback: state bad")
	}

	escPath := request.URL.EscapedPath()
	sep := strings.Split(escPath, "/")
	domain := strings.ToLower(sep[len(sep)-1])

	data, err := getUserDataFromGoogle(request.FormValue("code"), dc, domain)
	if err != nil {
		Debugfln("OauthGoogleCallback:err: %#v", err)

		return "", fmt.Errorf("ERROR: OauthGoogleCallback: code bad - %s", err.Error())
	}

	Debugfln("OauthGoogleCallback:2: No error...")

	if len(data) > 0 {

		return string(data), nil

	} else {

		err = errors.New("unable to get Google profile")
		Debugfln("OauthGoogleCallback:err: %#v", err)
		return "", err

	}
}

func GenerateStateOauthCookie(resp *http.Response) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	Debugfln("GenerateStateOauthCookie:1: Adding cookie with time until: %s", expiration.String())

	b, err := u.GenerateRandomBytes(16, true)
	if err != nil {
		panic(fmt.Sprintf("generateRandomBytes is unavailable: failed with %#v", err))
	}

	state := string(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}

	resp.Header.Add("Set-Cookie", cookie.String())

	return state
}

func getUserDataFromGoogle(code string, dc c.DomainConfig, emailDomain string) ([]byte, error) {
	// Use code to get token and get user info from Google.
	Debugfln("getUserDataFromGoogle:1: Starting...")

	setOauthConfig(dc, emailDomain)

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		Debugfln("getUserDataFromGoogle:1:err: %#v", err)
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	Debugfln("getUserDataFromGoogle:2: Trying get google profile...")

	client := googleOauthConfig.Client(oauth2.NoContext, token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	Debugfln("getUserDataFromGoogle:2: Userinfo status code: %d", response.StatusCode)

	if err != nil {
		Debugfln("getUserDataFromGoogle:2:err: %#v", err)
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	Debugfln("getUserDataFromGoogle:3: Trying to return google profile...")

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Debugfln("getUserDataFromGoogle:3:err: %#v", err)
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	} else {
		Debugfln("getUserDataFromGoogle:3: Response length: %d", len(contents))
	}

	defer response.Body.Close()

	Debugfln("getUserDataFromGoogle:4: Returning google profile")

	return contents, nil
}
