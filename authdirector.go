package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

func AuthIDPDirector(requ *http.Request, resp *http.Response) error {
	if requ.Method != "POST" {
		return errors.New("Incorrect method")
	}

	email := requ.Form.Get("email")

	if email == "" {
		return errors.New("No email set")
	}

	if strings.HasSuffix(email, "@digital.cabinet-office.gov.uk") {
		OAuthGoogleLogin(resp)
		return nil
	} else {
		return errors.New("Unknown domain")
	}
}

func AuthRequestDecision(request *http.Request) (*http.Response, error) {
	res := &http.Response{}
	res.Header = http.Header{}

	if request.URL.Path == "/auth/login" && request.Method == "GET" {

		t, err := template.ParseFiles("auth/login.html")
		if err != nil {
			return HTTPErrorResponse(err), err
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, nil); err != nil {
			return HTTPErrorResponse(err), err
		}

		res = &http.Response{
			Status:     "OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(tpl.Bytes())),
		}
		return res, nil

	} else if request.URL.Path == "/auth/login" && request.Method == "POST" {

		err := AuthIDPDirector(request, res)
		if err != nil {
			return HTTPErrorResponse(err), err
		}
		return res, nil

	} else if request.URL.Path == "/auth/google/callback" {

		OauthGoogleCallback(request, res)
		return res, nil

	}

	return res, nil
}
