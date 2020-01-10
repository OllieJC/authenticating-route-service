package google_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "authenticating-route-service/internal/configurator"
	g "authenticating-route-service/internal/google"
)

var _ = Describe("Google", func() {
	It("should return a string and cookie with GenerateStateOauthCookie", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		r := &http.Response{}
		r.Header = http.Header{}
		cookieStr := g.GenerateStateOauthCookie(r)

		data, err := base64.StdEncoding.DecodeString(cookieStr)

		Expect(err).NotTo(HaveOccurred())
		Expect(data).Should(HaveLen(16))
	})

	It("should attempt to query Google", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		response := &http.Response{}
		response.Header = http.Header{}

		cookieStr := g.GenerateStateOauthCookie(response)

		rcookie := response.Header.Get("Set-Cookie")

		path := "/auth/callback/google/email.example.local"
		request, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)
		request.Header = http.Header{"Cookie": []string{rcookie}}
		request.Form = url.Values{}
		request.Form.Add("state", cookieStr)
		request.Form.Add("code", "xxx")

		dc := c.DomainConfig{Domain: "example.local"}
		cbResp, err := g.OauthGoogleCallback(request, response, dc)

		Expect(err).To(HaveOccurred())
		Expect(cbResp).To(Equal(""))
	})

	It("should set an oauthstate cookie, location header and redirect with OAuthGoogleLogin", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		const (
			expectedHostnameInRedirect = "accounts.google.com"
		)

		dc := c.DomainConfig{Domain: "example.local"}

		r := &http.Response{}
		r.Header = http.Header{}
		g.OAuthGoogleLogin(r, dc, "email.example.local")

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
