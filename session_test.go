package main_test

import (

	//"github.com/jarcoal/httpmock"

	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
)

var _ = Describe("Sessions", func() {
	It("should not error when encrypting and decrypting", func() {
		var testString = "Testing123."
		const testPassphrase = "PassphraseHere!"

		encString, err := s.Encrypt(testString, testPassphrase)
		Expect(err).NotTo(HaveOccurred())

		decString, err := s.Decrypt(encString, testPassphrase)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(decString)).To(Equal(testString))
	})

	It("should add a cookie with 6 hour expiration to response", func() {

	})

	It("should check a cookie value in a request", func() {
		const (
			testString = "Testing123."
		)

		sess := s.NewCustomSession()
		sess.Provider = "Test"
		sess.UserData = testString
		b, err := json.Marshal(sess)
		Expect(err).NotTo(HaveOccurred())

		encString, err := s.Encrypt(string(b), "")
		Expect(err).NotTo(HaveOccurred())

		cookie := http.Cookie{Name: "_session", Value: encString}
		request := &http.Request{Header: http.Header{"Cookie": []string{cookie.String()}}}

		bTest := s.CheckCookie(request)
		Expect(bTest).To(BeTrue())
	})
})
