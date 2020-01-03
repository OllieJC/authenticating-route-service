package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
	i "authenticating-route-service/internal"
)

var _ = Describe("Main", func() {
	os.Setenv("DOMAIN_CONFIG_FILEPATH", "test/data/example.yml")

	It("should respond to a backing service with the 'X-Cf-Forwarded-Url' set and a cookie", func() {
		const (
			expected          = "hi"
			skipSslValidation = false
			sigHeader         = "X-CF-Proxy-Signature"
			metaHeader        = "X-CF-Proxy-Metadata"
			expectedSig       = "aaaaa"
			expectedMeta      = "bbbbb"
		)

		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(expected))
		}))
		defer backend.Close()

		_, err := url.Parse(backend.URL)
		Expect(err).NotTo(HaveOccurred())

		roundTripper := s.NewAuthRoundTripper(skipSslValidation)
		proxyHandler := s.NewProxy(roundTripper)

		frontend := httptest.NewServer(proxyHandler)
		defer frontend.Close()

		req, _ := http.NewRequest("GET", frontend.URL, nil)
		req.Header.Add("X-Cf-Forwarded-Url", backend.URL)
		req.Header.Add(sigHeader, expectedSig)
		req.Header.Add(metaHeader, expectedMeta)

		sess := i.NewCustomSession()
		sess.Provider = "Test"
		sess.UserData = "abc123"
		b, err := json.Marshal(sess)
		Expect(err).NotTo(HaveOccurred())

		encString, err := i.Encrypt(string(b), "")
		Expect(err).NotTo(HaveOccurred())

		cookie := &http.Cookie{Name: "_session", Value: encString, Expires: time.Now().Add(6 * time.Hour)}
		req.AddCookie(cookie)

		req.Close = true
		res, err := frontend.Client().Do(req)

		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).To(Equal(expected))

		Expect(len(res.Header)).Should(BeNumerically(">", 0))

		Expect(res.Header.Get(sigHeader)).To(Equal(expectedSig))
		Expect(res.Header.Get(metaHeader)).To(Equal(expectedMeta))
	})

	It("should return bad email when post to /auth/login", func() {
		const (
			notExpectedBody   = "hello"
			expectedBody      = "govuk-input--error"
			sigHeader         = "X-CF-Proxy-Signature"
			metaHeader        = "X-CF-Proxy-Metadata"
			expectedSig       = "aaaaa"
			expectedMeta      = "bbbbb"
			skipSslValidation = false
		)

		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(notExpectedBody))
		}))
		backend.URL = "http://example.com"
		defer backend.Close()

		_, err := url.Parse(backend.URL)
		Expect(err).NotTo(HaveOccurred())

		roundTripper := s.NewAuthRoundTripper(skipSslValidation)
		proxyHandler := s.NewProxy(roundTripper)

		frontend := httptest.NewServer(proxyHandler)
		defer frontend.Close()

		reqFeUrl := fmt.Sprintf("%s/auth/login", frontend.URL)
		reqBeUrl := fmt.Sprintf("%s/auth/login", backend.URL)

		fmt.Println("reqFeUrl:", reqFeUrl)
		fmt.Println("reqBeUrl:", reqBeUrl)

		req, _ := http.NewRequest("POST", reqFeUrl, nil)
		req.Header.Add("X-Cf-Forwarded-Url", reqBeUrl)
		req.Header.Add(sigHeader, expectedSig)
		req.Header.Add(metaHeader, expectedMeta)
		req.Close = true
		res, err := frontend.Client().Do(req)

		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).ToNot(ContainSubstring(notExpectedBody))
		Expect(string(bodyBytes)).To(ContainSubstring(expectedBody))

		Expect(len(res.Header)).Should(BeNumerically(">", 0))

		Expect(res.Header.Get(sigHeader)).To(Equal(expectedSig))
		Expect(res.Header.Get(metaHeader)).To(Equal(expectedMeta))
	})

	It("should return bad email when post to /auth/google/callback", func() {
		const (
			notExpectedBody   = "hello"
			expectedBody      = "state bad"
			sigHeader         = "X-CF-Proxy-Signature"
			metaHeader        = "X-CF-Proxy-Metadata"
			expectedSig       = "aaaaa"
			expectedMeta      = "bbbbb"
			skipSslValidation = false
		)

		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(notExpectedBody))
		}))
		backend.URL = "http://example.com"
		defer backend.Close()

		_, err := url.Parse(backend.URL)
		Expect(err).NotTo(HaveOccurred())

		roundTripper := s.NewAuthRoundTripper(skipSslValidation)
		proxyHandler := s.NewProxy(roundTripper)

		frontend := httptest.NewServer(proxyHandler)
		defer frontend.Close()

		response := &http.Response{}
		response.Header = http.Header{}
		cookieStr := i.GenerateStateOauthCookie(response)
		rcookie := response.Header.Get("Set-Cookie")

		reqFeUrl := fmt.Sprintf("%s/auth/google/callback", frontend.URL)
		reqBeUrl := fmt.Sprintf("%s/auth/google/callback", backend.URL)

		req, _ := http.NewRequest("POST", reqFeUrl, nil)
		req.Header.Add("X-Cf-Forwarded-Url", reqBeUrl)
		req.Header.Add(sigHeader, expectedSig)
		req.Header.Add(metaHeader, expectedMeta)

		req.Header.Add("Cookie", rcookie)

		req.Form = url.Values{}
		req.Form.Add("state", cookieStr)

		req.Close = true
		res, err := frontend.Client().Do(req)

		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).ToNot(ContainSubstring(notExpectedBody))
		Expect(string(bodyBytes)).To(ContainSubstring(expectedBody))

		Expect(len(res.Header)).Should(BeNumerically(">", 0))

		Expect(res.Header.Get(sigHeader)).To(Equal(expectedSig))
		Expect(res.Header.Get(metaHeader)).To(Equal(expectedMeta))
	})
})
