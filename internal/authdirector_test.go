package internal_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal"
)

var _ = Describe("AuthDirector", func() {

	Context("AuthIDPDirector", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		req, _ := http.NewRequest("POST", "http://example.local/auth/non-exist", nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp := &http.Response{}
		resp.Header = http.Header{}

		It("should return an error with no form data in post request", func() {
			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("Email address not recognised"))
		})

		It("should return an error with an invalid email domain", func() {
			req.PostForm = url.Values{
				"email": {"test@invalid.uk"},
			}
			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("Email address not recognised"))
		})

		It("should return an redirect with a valid email domain", func() {

			req.PostForm = url.Values{
				"email": {"test@email.example.local"},
			}
			err := s.AuthIDPDirector(req, resp)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
		})

		It("should return an error if not a post request", func() {
			req.Method = "GET"

			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("Incorrect method"))
		})
	})

	Context("AuthRequestDecision", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")

		It("should return an login page when get '/auth/login'", func() {
			const (
				path      = "/auth/login"
				expected  = `id="email"`
				expected2 = `<link href="/auth/assets/all.min.css" rel="stylesheet" />`
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(bodyBytes)).To(ContainSubstring(expected))
			Expect(string(bodyBytes)).To(ContainSubstring(expected2))
		})

		It("should return a redirect when post '/auth/login'", func() {
			const (
				path        = "/auth/login"
				notexpected = "govuk-input--error"
			)

			var err error

			req, _ := http.NewRequest("POST", fmt.Sprintf("http://example.local%s", path), nil)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.PostForm = url.Values{
				"email": {"test@email.example.local"},
			}

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).ToNot(ContainSubstring(notexpected))
		})

		It("should return a CSS doc when get '/auth/assets/all.min.css'", func() {
			const (
				path                = "/auth/assets/all.min.css"
				expectedContent     = ".govuk-link"
				expectedContentType = "text/css"
				expectedStatusCode  = 200
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).Should(Equal(expectedStatusCode))

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring(expectedContent))

			Expect(resp.Header.Get("Content-Type")).To(Equal(expectedContentType))
		})

		It("should return a JS doc when get '/auth/assets/all.js'", func() {
			const (
				path                = "/auth/assets/all.js"
				expectedContent     = "function("
				expectedContentType = "application/javascript"
				expectedStatusCode  = 200
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).Should(Equal(expectedStatusCode))

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring(expectedContent))

			Expect(resp.Header.Get("Content-Type")).To(Equal(expectedContentType))
		})

		It("should return an image when get '/auth/assets/images/govuk-crest.png'", func() {
			const (
				path                = "/auth/assets/images/govuk-crest.png"
				expectedContentType = "image/png"
				expectedStatusCode  = 200
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).Should(Equal(expectedStatusCode))

			Expect(resp.Body).NotTo(BeNil())
			_, err = ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.Header.Get("Content-Type")).To(Equal(expectedContentType))
		})

		It("should return a 404 when get '/auth/assets/not-exist'", func() {
			const (
				path               = "/auth/assets/not-exist"
				expectedStatusCode = 404
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(expectedStatusCode))
		})

		It("should set-cookie to a deleted cookie with /auth/logout", func() {
			const (
				path               = "/auth/logout"
				expectedStatusCode = http.StatusSeeOther
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.local%s", path), nil)

			resp, err := s.AuthRequestDecision(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(expectedStatusCode))

			cookieRaw := resp.Header.Get("Set-Cookie")
			Expect(cookieRaw).Should(ContainSubstring("_session"))
			Expect(cookieRaw).Should(ContainSubstring("Max-Age=0"))

			expectedYear := strconv.Itoa(time.Now().AddDate(-1, -1, -1).Year())
			Expect(cookieRaw).Should(ContainSubstring(expectedYear))
		})
	})
})
