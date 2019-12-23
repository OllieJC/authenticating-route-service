package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
)

var _ = Describe("AuthDirector", func() {
	Context("AuthIDPDirector", func() {

		req, _ := http.NewRequest("POST", "http://localhost:8080/auth/non-exist", nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp := &http.Response{}
		resp.Header = http.Header{}

		It("should return an error with no form data in post request", func() {
			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("No email set"))
		})

		It("should return an error with an invalid email domain", func() {
			req.Form = url.Values{
				"email": {"test@invalid.uk"},
			}
			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("Unknown domain"))
		})

		It("should return an redirect with a valid email domain", func() {

			req.Form = url.Values{
				"email": {"test@digital.cabinet-office.gov.uk"},
			}
			err := s.AuthIDPDirector(req, resp)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusTemporaryRedirect))
		})

		It("should return an error if not a post request", func() {
			req.Method = "GET"

			err := s.AuthIDPDirector(req, resp)

			Expect(err).Should(MatchError("Incorrect method"))
		})
	})

	Context("AuthRequestDecision", func() {
		It("should return an login page when get '/auth/login'", func() {
			const (
				path     = "/auth/login"
				expected = "login"
			)

			var err error

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080%s", path), nil)

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring(expected))
		})

		It("should return a redirect when post '/auth/login'", func() {
			const (
				path        = "/auth/login"
				notexpected = "login"
			)

			var err error

			req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080%s", path), nil)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Form = url.Values{
				"email": {"test@digital.cabinet-office.gov.uk"},
			}

			resp, err := s.AuthRequestDecision(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.Body).NotTo(BeNil())
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).ToNot(ContainSubstring(notexpected))
		})
	})
})
