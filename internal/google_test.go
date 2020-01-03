package internal_test

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal"
	c "authenticating-route-service/internal/configurator"
)

var _ = Describe("Google", func() {
	It("should return a string and cookie with GenerateStateOauthCookie", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		r := &http.Response{}
		r.Header = http.Header{}
		cookieStr := s.GenerateStateOauthCookie(r)

		data, err := base64.StdEncoding.DecodeString(cookieStr)

		Expect(err).NotTo(HaveOccurred())
		Expect(data).Should(HaveLen(16))
	})

	It("should attempt to query Google", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		response := &http.Response{}
		response.Header = http.Header{}

		cookieStr := s.GenerateStateOauthCookie(response)

		rcookie := response.Header.Get("Set-Cookie")
		request := &http.Request{Header: http.Header{"Cookie": []string{rcookie}}}

		request.Form = url.Values{}
		request.Form.Add("state", cookieStr)
		request.Form.Add("code", "xxx")

		dc := c.DomainConfig{Domain: "example.com"}
		cbResp, err := s.OauthGoogleCallback(request, response, dc)

		Expect(err).To(HaveOccurred())
		Expect(cbResp).To(Equal(""))
	})

	It("should set an oauthstate cookie, location header and redirect with OAuthGoogleLogin", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		const (
			expectedHostnameInRedirect = "accounts.google.com"
		)

		dc := c.DomainConfig{Domain: "example.com"}

		r := &http.Response{}
		r.Header = http.Header{}
		s.OAuthGoogleLogin(r, dc)

		Expect(r.StatusCode).To(Equal(http.StatusSeeOther))

		rcookie := r.Header.Get("Set-Cookie")
		request := &http.Request{Header: http.Header{"Cookie": []string{rcookie}}}

		cookie, err := request.Cookie("oauthstate")
		Expect(err).NotTo(HaveOccurred())

		data, err := base64.StdEncoding.DecodeString(cookie.Value)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).Should(HaveLen(16))

		url, err := r.Location()
		Expect(err).NotTo(HaveOccurred())
		Expect(url.Hostname()).Should(Equal(expectedHostnameInRedirect))

		bodyBytes, err := ioutil.ReadAll(r.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).Should(ContainSubstring(`http-equiv="refresh"`))
		Expect(string(bodyBytes)).Should(ContainSubstring(expectedHostnameInRedirect))
	})
})
