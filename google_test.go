package main_test

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
)

var _ = Describe("Google", func() {
	It("should return a string and cookie with GenerateStateOauthCookie", func() {
		r := &http.Response{}
		r.Header = http.Header{}
		cookieStr := s.GenerateStateOauthCookie(r)

		data, err := base64.StdEncoding.DecodeString(cookieStr)

		Expect(err).NotTo(HaveOccurred())
		Expect(data).Should(HaveLen(16))
	})

	It("should attempt to query Google", func() {
		response := &http.Response{}
		response.Header = http.Header{}

		cookieStr := s.GenerateStateOauthCookie(response)

		rcookie := response.Header.Get("Set-Cookie")
		request := &http.Request{Header: http.Header{"Cookie": []string{rcookie}}}

		request.Form = url.Values{}
		request.Form.Add("state", cookieStr)
		request.Form.Add("code", "4/uwEID0d1_BYHEE5UkjsFW08bKwhCXw5U_ajdadKBoygAwUKhT4_YSLO6pR6CO5UJFbw5m8s7DeGiKylK3Ugb1mk")

		cbResp, err := s.OauthGoogleCallback(request, response)

		Expect(err).To(HaveOccurred())
		Expect(cbResp).To(Equal(""))
	})

	It("should set an oauthstate cookie, location header and redirect with OAuthGoogleLogin", func() {
		const (
			expectedHostnameInRedirect = "accounts.google.com"
		)

		r := &http.Response{}
		r.Header = http.Header{}
		s.OAuthGoogleLogin(r)

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
