package main_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
)

var _ = Describe("Main", func() {
	It("should respond to a backing service with the 'X-Cf-Forwarded-Url' set", func() {
		const (
			expected          = "hi"
			skipSslValidation = false
		)

		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(expected))
		}))
		defer backend.Close()

		_, err := url.Parse(backend.URL)
		Expect(err).NotTo(HaveOccurred())

		roundTripper := s.NewLoggingRoundTripper(skipSslValidation)
		proxyHandler := s.NewProxy(roundTripper, skipSslValidation)

		frontend := httptest.NewServer(proxyHandler)
		defer frontend.Close()

		req, _ := http.NewRequest("GET", frontend.URL, nil)
		req.Header.Add("X-Cf-Forwarded-Url", backend.URL)
		req.Close = true
		res, err := frontend.Client().Do(req)

		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()

		bodyBytes, err := ioutil.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).To(Equal(expected))
	})
})
