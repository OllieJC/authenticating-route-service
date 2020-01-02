package internal

import (
	. "authenticating-route-service/pkg/debugprint"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	errBadMethod error = errors.New("Incorrect method")
	errBadEmail  error = errors.New("Email address not recognised")
)

func AuthIDPDirector(request *http.Request, response *http.Response) error {

	Debugfln("AuthIDPDirector:1: Request type: %s", request.Method)

	if request.Method != "POST" {
		return errBadMethod
	}

	email := request.PostFormValue("email")

	if email == "" {
		return errBadEmail
	}

	if strings.HasSuffix(email, "@digital.cabinet-office.gov.uk") {
		OAuthGoogleLogin(response)

		Debugfln("AuthIDPDirector:2: Returning good email.")

		return nil
	}

	Debugfln("AuthIDPDirector:2: Returning bad email.")

	return errBadEmail
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

	Debugfln("returnAsset:1: Start return.")

	valFile := regexp.MustCompile(`^/auth/assets/(?P<file>(?:(?:fonts|images)/)?[\w\.-]+\.[\w\.-]+)`)
	extFile := valFile.FindStringSubmatch(request.URL.Path)

	if len(extFile) == 2 && extFile[1] != "" {

		Debugfln("returnAsset:2: Getting %s.", extFile)

		fStr := filepath.Join(StaticAssetPath, filepath.Clean(extFile[1]))

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
	return HTTPNotFoundResponse(err), nil
}

func AuthRequestDecision(request *http.Request) (*http.Response, error) {

	Debugfln("AuthRequestDecision:1: Starting...")

	response := EmptyHTTPResponse(request)
	var err error

	if strings.HasPrefix(request.URL.Path, "/auth/status") {

		body := []byte("false")
		if ok, _ := CheckCookie(request); ok {
			body = []byte("true")
		}
		response.StatusCode = http.StatusOK
		response.Body = ioutil.NopCloser(bytes.NewReader(body))

	} else if strings.HasPrefix(request.URL.Path, "/auth/assets") && request.Method == "GET" {

		Debugfln("AuthRequestDecision:2: Asset")

		return returnAsset(request)

	} else if request.URL.Path == "/auth/login" && request.Method == "GET" {

		Debugfln("AuthRequestDecision:3: GET /auth/login")

		tpd := NewTemplatePageData()
		tpd.Title = "Login"
		response, err = TemplateResponse("login.html", http.StatusOK, tpd)
		if err != nil {
			return HTTPErrorResponse(err), err
		}

	} else if request.URL.Path == "/auth/logout" {

		Debugfln("AuthRequestDecision:3: GET /auth/logout")

		RemoveCookie(response)
		RedirectResponse(response, http.StatusSeeOther, "/auth/login")

	} else if request.URL.Path == "/auth/login" && request.Method == "POST" {

		Debugfln("AuthRequestDecision:4: POST /auth/login")

		err = AuthIDPDirector(request, response)

		if err == errBadEmail {
			tpd := NewTemplatePageData()
			tpd.Title = "Bad Email"
			response, err = TemplateResponse("bad-email.html", http.StatusUnauthorized, tpd)
		}

		if err != nil {
			Debugfln("AuthRequestDecision:4:err: %s", err.Error())

			return HTTPErrorResponse(err), err
		}

	} else if request.URL.Path == "/auth/google/callback" {

		Debugfln("AuthRequestDecision:5: /auth/google/callback")

		cbResp, err := OauthGoogleCallback(request, response)

		if err != nil {
			Debugfln("AuthRequestDecision:5:err: %s", err.Error())

			return HTTPErrorResponse(err), err
		}

		AddCookie(request, response, "Google", cbResp)
		RedirectResponse(response, http.StatusSeeOther, "/")

	} else {
		Debugfln("AuthRequestDecision:6: Response not found")

		response = HTTPNotFoundResponse(nil)
	}

	Debugfln("AuthRequestDecision:7: Returning response")

	return response, nil
}
