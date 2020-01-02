package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	_securityOption map[string]string
)

const noSetSecOpt = "NO-SET"

func securityOption(option string) (string, string) {
	var val string

	if option == "" {
		return option, val
	}

	optEnvVar := strings.Replace(strings.ToUpper(option), "-", "_", -1)
	eso := fmt.Sprintf("ENV_SEC_OPT_%s", optEnvVar)

	if InternalTest {
		val = os.Getenv(eso)
	} else {
		if _securityOption == nil {
			_securityOption = make(map[string]string)
		}

		var ok bool
		val, ok = _securityOption[optEnvVar]

		if ok == false {
			val = os.Getenv(eso)
			_securityOption[optEnvVar] = val
		}
	}

	return option, val
}

type templatePageData struct {
	Title      string
	ErrorText  string
	AssetsPath string
}

func NewTemplatePageData() templatePageData {
	t := templatePageData{}
	t.Title = ""
	t.ErrorText = ""
	t.AssetsPath = "/auth/assets"
	return t
}

func EmptyHTTPResponse(request *http.Request) (response *http.Response) {
	return &http.Response{
		Status:        "Not Implemented",
		StatusCode:    501,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          nil,
		ContentLength: 0,
		Request:       request,
		Header:        make(http.Header, 0),
	}
}

func TemplateResponse(templateFileName string, responseCode int, tpd templatePageData) (*http.Response, error) {
	response := EmptyHTTPResponse(nil)

	t, err := template.ParseGlob(filepath.Join(TemplatePath, "*.html"))
	if err != nil {
		log.Fatalf("Cannot parse templates: %#v", err)
		return nil, err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, templateFileName, tpd); err != nil {
		return nil, err
	}

	response.StatusCode = responseCode
	response.Body = ioutil.NopCloser(bytes.NewReader(tpl.Bytes()))
	response.Header.Add("Cache-Control", "max-age=60, public")

	return response, nil
}

func RedirectResponse(response *http.Response, status int, url string) {
	body := fmt.Sprintf(`<head>
                         <meta http-equiv="refresh" content="0; URL=%s" />
                       </head>`, url)

	response.Status = "Redirect"
	response.StatusCode = status
	response.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	response.Header.Add("Location", url)
}

func HTTPErrorResponse(err error) *http.Response {
	tpd := NewTemplatePageData()
	tpd.Title = "Error"
	tpd.ErrorText = err.Error()
	t, _ := TemplateResponse("error.html", http.StatusInternalServerError, tpd)
	return t
}

func HTTPNotFoundResponse(err error) *http.Response {
	tpd := NewTemplatePageData()
	tpd.Title = "Not Found"
	tpd.ErrorText = "Element not found"
	t, _ := TemplateResponse("error.html", http.StatusNotFound, tpd)
	return t
}

func AddSecurityHeaders(response *http.Response) {
	addSecurityHeader(response, "X-Xss-Protection", "1; mode=block")
	addSecurityHeader(response, "X-Content-Type-Options", "nosniff")
	addSecurityHeader(response, "X-Frame-Options", "DENY")
	addSecurityHeader(response, "Content-Security-Policy", "default-src 'self'")
	addSecurityHeader(response, "Referrer-Policy", "strict-origin-when-cross-origin")
	addSecurityHeader(response, "Feature-Policy", "vibrate 'none'; geolocation 'none'; microphone 'none'; camera 'none'; payment 'none'; notifications 'none';")
}

func addSecurityHeader(response *http.Response, header string, defaultStr string) {
	_, val := securityOption(header)
	if val == "" {
		response.Header.Add(header, defaultStr)
	} else if val != noSetSecOpt {
		response.Header.Add(header, val)
	}
}
