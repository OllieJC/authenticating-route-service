package main

import (
	i "authenticating-route-service/internal"
	c "authenticating-route-service/internal/configurator"
	h "authenticating-route-service/internal/httphelper"
	d "authenticating-route-service/pkg/debugprint"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort            = 8080
	cfForwardedURLHeader   = "X-Cf-Forwarded-Url"
	cfProxySignatureHeader = "X-Cf-Proxy-Signature"
	cfProxyMetadataHeader  = "X-CF-Proxy-Metadata"
	redirectCookieName     = "_redirectPath"
)

func main() {
	var (
		skipSslValidation bool
		port              int64
	)

	port, _ = strconv.ParseInt(os.Getenv("PORT"), 10, 16)
	if port == 0 {
		port = defaultPort
	}

	ssv := os.Getenv("SKIP_SSL_VALIDATION")
	if len(ssv) != 0 {
		skipSslValidation, _ = strconv.ParseBool(ssv)
	} else {
		skipSslValidation = true
	}

	log.SetOutput(os.Stdout)

	roundTripper := NewAuthRoundTripper(skipSslValidation)
	proxy := NewProxy(roundTripper)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), proxy))
}

// NewProxy sets up a http Handler using the custom AuthRoundTripper
func NewProxy(transport http.RoundTripper) http.Handler {
	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			forwardedURL := req.Header.Get(cfForwardedURLHeader)
			d.Debugfln("NewProxy:1: %s", forwardedURL)

			var body []byte
			var err error
			if req.Body != nil {
				body, err = ioutil.ReadAll(req.Body)
				if err != nil {
					log.Fatalln("NewProxy:err:", err.Error())
				}
				req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}

			// Note that url.Parse is decoding any url-encoded characters.
			url, err := url.Parse(forwardedURL)
			if err != nil {
				log.Fatalln("NewProxy:err:", err.Error())
			}

			req.URL = url
			req.Host = url.Host
		},
		Transport: transport,
	}
	return reverseProxy
}

// AuthRoundTripper object, exported for use in tests
type AuthRoundTripper struct {
	transport http.RoundTripper
}

// NewAuthRoundTripper returns an AuthRoundTripper
func NewAuthRoundTripper(skipSslValidation bool) *AuthRoundTripper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSslValidation},
	}
	return &AuthRoundTripper{
		transport: tr,
	}
}

// RoundTrip returns a response and error from a request
func (lrt *AuthRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	path := request.URL.EscapedPath()

	d.Debugfln("RoundTrip:1: path: %s", path)

	if strings.HasPrefix(path, "/auth") {

		d.Debugfln("RoundTrip:2: Auth request.")

		response, err = i.AuthRequestDecision(request)
		if err != nil {
			response = h.HTTPErrorResponse(err)
		}

	} else {

		var doBackEndRequest bool

		unauthPath := c.IsUnauthPath(request)
		d.Debugfln("RoundTrip:1: unauthPath: %t", unauthPath)

		if unauthPath {
			doBackEndRequest = true
		} else {
			doBackEndRequest, _ = i.CheckCookie(request)
			d.Debugfln("RoundTrip:1: CheckCookie: %t", doBackEndRequest)
		}

		if doBackEndRequest {

			d.Debugfln("RoundTrip:2: Forwarding to: %s", request.URL.String())

			response, err = lrt.transport.RoundTrip(request)
			if err != nil {
				response = h.HTTPErrorResponse(err)
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				response = h.HTTPErrorResponse(err)
			} else {
				response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}

			i.AddCookie(request, response, "", "")

			if response.Header.Get("Cache-Control") == "" {
				response.Header.Add("Cache-Control", "max-age=1, private")
			}

		} else {

			d.Debugfln("RoundTrip:2: Redirecting to login page")
			response = h.EmptyHTTPResponse(request)

			if true { //c.IsPathRedirectionEnabled(request) {
				d.Debugfln("RoundTrip:2: Add redirect cookie")

				cookie := &http.Cookie{
					Name:     "_redirectPath",
					Value:    request.URL.RequestURI(),
					Expires:  time.Now().Add(2 * time.Hour),
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
				}
				response.Header.Add("Set-Cookie", cookie.String())
			}

			h.RedirectResponse(response, http.StatusSeeOther, "/auth/login")

		}
	}

	if request.Header.Get(redirectCookieName) != "" {
		h.RemoveCookie(response, redirectCookieName)
	}

	sigHeader := request.Header.Get(cfProxySignatureHeader)
	metaHeader := request.Header.Get(cfProxyMetadataHeader)

	response.Header.Add(cfProxySignatureHeader, sigHeader)
	response.Header.Add(cfProxyMetadataHeader, metaHeader)

	h.AddSecurityHeaders(request, response)

	d.Debugfln("RoundTrip:3: Responding...")

	return response, nil
}
