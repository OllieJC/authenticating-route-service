package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
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

func returnAsset(request *http.Request) (*http.Response, error) {
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
				if err != nil {
					return HTTPErrorResponse(err), err
				}
			}

			dat, err := ioutil.ReadFile(fStr)
			if err == nil {
				res := &http.Response{
					Status:     "OK",
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader(dat)),
				}
				res.Header = http.Header{}
				res.Header.Add("Content-Type", contentType)
				return res, nil
			}
		}
	}

	err = errors.New("No asset found with that filename")
	return HTTPNotFoundResponse(err), err
}

func AuthRequestDecision(request *http.Request) (*http.Response, error) {
	res := &http.Response{}
	res.Header = http.Header{}

	if strings.HasPrefix(request.URL.Path, "/auth/assets") && request.Method == "GET" {
		return returnAsset(request)

	} else if request.URL.Path == "/auth/login" && request.Method == "GET" {

		t, err := template.ParseGlob("./templates/*.html")
		if err != nil {
			log.Println("Cannot parse templates:", err)
			return HTTPErrorResponse(err), err
		}

		var tpl bytes.Buffer
		apd := authPageData{
			"Test",
			"/auth/assets",
		}
		if err := t.ExecuteTemplate(&tpl, "login.html", apd); err != nil {
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
