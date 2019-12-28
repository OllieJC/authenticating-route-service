package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type authPageData struct {
	Title      string
	AssetsPath string
}

func AuthIDPDirector(request *http.Request, response *http.Response) error {
	if request.Method != "POST" {
		return errors.New("Incorrect method")
	}

	email := request.Form.Get("email")

	if email == "" {
		return errors.New("No email set")
	}

	if strings.HasSuffix(email, "@digital.cabinet-office.gov.uk") {
		OAuthGoogleLogin(response)
		return nil
	} else {
		return errors.New("Unknown domain")
	}
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getFileContentType(fStr string) (string, error) {
	var (
		out *os.File
		err error
	)

	out, err = os.Open(fStr)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func returnAsset(request *http.Request, response *http.Response) error {
	var err error

	valFile := regexp.MustCompile(`^/auth/assets/(?P<file>(?:(?:fonts|images)/)?[\w\.-]+\.[\w\.-]+)`)
	extFile := valFile.FindStringSubmatch(request.URL.Path)

	if len(extFile) == 2 && extFile[1] != "" {
		fStr := filepath.Join("templates/assets", filepath.Clean(extFile[1]))

		if exists(fStr) {
			var contentType string

			if strings.HasSuffix(fStr, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(fStr, ".js") {
				contentType = "application/javascript"
			} else {
				// Get the content
				contentType, err = getFileContentType(fStr)
			}

			dat, err := ioutil.ReadFile(fStr)
			if err == nil {
				response.Status = "200 OK"
				response.StatusCode = 200
				response.Body = ioutil.NopCloser(bytes.NewReader(dat))
				response.Header.Add("Content-Type", contentType)
				return nil
			}
		}
	}

	if err == nil {
		response.Status = "Not Found"
		response.StatusCode = 404
		response.Body = nil
	}

	return err
}

func AuthRequestDecision(request *http.Request) (*http.Response, error) {
	var err error

	response := EmptyHTTPResponse(request)

	if strings.HasPrefix(request.URL.Path, "/auth/assets") && request.Method == "GET" {

		err = returnAsset(request, response)
		if err != nil {
			return nil, err
		}
		return response, nil

	} else if request.URL.Path == "/auth/login" && request.Method == "GET" {

		t, err := template.ParseGlob("./templates/*.html")
		if err != nil {
			return nil, err
		}

		var tpl bytes.Buffer
		apd := authPageData{
			"Test",
			"/auth/assets",
		}
		if err := t.ExecuteTemplate(&tpl, "login.html", apd); err != nil {
			return nil, err
		}

		response.Status = "200 OK"
		response.StatusCode = 200
		response.Body = ioutil.NopCloser(bytes.NewReader(tpl.Bytes()))
		return response, nil

	} else if request.URL.Path == "/auth/login" && request.Method == "POST" {

		err := AuthIDPDirector(request, response)
		if err != nil {
			return nil, err
		}
		return response, nil

	} else if request.URL.Path == "/auth/google/callback" {

		OauthGoogleCallback(request, response)
		return response, nil

	}

	return nil, nil
}
