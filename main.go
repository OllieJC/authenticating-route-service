package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	DEFAULT_PORT              = 8080
	CF_FORWARDED_URL_HEADER   = "X-Cf-Forwarded-Url"
	CF_PROXY_SIGNATURE_HEADER = "X-Cf-Proxy-Signature"
	CF_PROXY_METADATA_HEADER  = "X-CF-Proxy-Metadata"
)

var DebugOut io.Writer = ioutil.Discard

func main() {
	var (
		skipSslValidation bool
		port              int64
	)

	port, _ = strconv.ParseInt(os.Getenv("PORT"), 10, 16)
	if port == 0 {
		port = DEFAULT_PORT
	}

	ssv := os.Getenv("SKIP_SSL_VALIDATION")
	if len(ssv) != 0 {
		skipSslValidation, _ = strconv.ParseBool(ssv)
	} else {
		skipSslValidation = true
	}

	log.SetOutput(os.Stdout)

	roundTripper := NewLoggingRoundTripper(skipSslValidation)
	proxy := NewProxy(roundTripper, skipSslValidation)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), proxy))
}

func NewProxy(transport http.RoundTripper, skipSslValidation bool) http.Handler {
	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			forwardedURL := req.Header.Get(CF_FORWARDED_URL_HEADER)
			debug("NewProxy:1: %s\n", forwardedURL)

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

type LoggingRoundTripper struct {
	transport http.RoundTripper
}

func NewLoggingRoundTripper(skipSslValidation bool) *LoggingRoundTripper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSslValidation},
	}
	return &LoggingRoundTripper{
		transport: tr,
	}
}

func (lrt *LoggingRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {

	path := request.URL.EscapedPath()

	debug("RoundTrip:1: %s\n", request.URL)

	if strings.HasPrefix(path, "/auth") {

		debug("RoundTrip:2: Auth request to: %s\n", request.URL.String())

		response, err = AuthRequestDecision(request)
		if err != nil {
			response = HTTPErrorResponse(err)
		}

	} else {

		if CheckCookie(request) {

			debug("RoundTrip:2: Forwarding to: %s", request.URL.String())

			response, err = lrt.transport.RoundTrip(request)
			if err != nil {
				response = HTTPErrorResponse(err)
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				response = HTTPErrorResponse(err)
			} else {
				response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}

		} else {

			debug("RoundTrip:2: Redirecting to login page")
			response = EmptyHTTPResponse(request)
			RedirectResponse(response, http.StatusSeeOther, "/auth/login")

		}
	}

	sigHeader := request.Header.Get(CF_PROXY_SIGNATURE_HEADER)
	metaHeader := request.Header.Get(CF_PROXY_METADATA_HEADER)

	response.Header.Add(CF_PROXY_SIGNATURE_HEADER, sigHeader)
	response.Header.Add(CF_PROXY_METADATA_HEADER, metaHeader)

	AddSecurityHeaders(response)

	debug("RoundTrip:3: Responding...")

	return response, nil
}
