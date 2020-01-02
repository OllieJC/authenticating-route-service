package internal_test

import (

	//"github.com/jarcoal/httpmock"

	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal"
)

func parseCookieTime(rawCookieStr string) (time.Time, error) {
	if rawCookieStr != "" && strings.Contains(rawCookieStr, ";") {
		splitCookie := strings.Split(rawCookieStr, ";")
		if len(splitCookie) >= 2 {
			expiryRaw := splitCookie[2]
			if strings.Contains(expiryRaw, "Expires") {
				expiryStr := strings.Split(expiryRaw, "=")[1]
				return time.Parse(time.RFC1123, expiryStr)
			}
		}
	}
	return time.Unix(0, 0), errors.New("Bad string")
}

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

	It("should add a cookie with AddCookie", func() {

		request := &http.Request{Header: http.Header{}}
		response := s.EmptyHTTPResponse(request)

		s.AddCookie(request, response, "Test", "abc123")

		cookieRawVal := response.Header.Get("Set-Cookie")
		Expect(cookieRawVal).ToNot(BeNil())

		t, err := parseCookieTime(cookieRawVal)
		Expect(err).NotTo(HaveOccurred())

		// cookie time should not equal the default time
		Expect(t.Unix()).ShouldNot(BeEquivalentTo(-62135596800))

		// cookie time should be greater than current time
		Expect(t.Unix()).Should(BeNumerically(">", time.Now().Unix()))

		// cookie time should be less than seven hours
		Expect(t.Unix()).Should(BeNumerically("<", time.Now().Add(7*time.Hour).Unix()))
	})

	It("should set expiry in past with RemoveCookie", func() {
		response := s.EmptyHTTPResponse(nil)

		s.RemoveCookie(response)

		cookieRawVal := response.Header.Get("Set-Cookie")
		Expect(cookieRawVal).ToNot(BeNil())

		t, err := parseCookieTime(cookieRawVal)
		Expect(err).NotTo(HaveOccurred())

		// cookie time should not equal the default time
		Expect(t.Unix()).ShouldNot(BeEquivalentTo(-62135596800))

		// cookie time should be less than current time
		Expect(t.Unix()).Should(BeNumerically("<", time.Now().Unix()))
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

		bTest, retSess := s.CheckCookie(request)
		Expect(bTest).To(BeTrue())
		Expect(retSess).Should(Equal(sess))
	})

	It("should renew cookie if set already in a request", func() {
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

		response := s.EmptyHTTPResponse(nil)
		s.AddCookie(request, response, "", "")

		cookieInResponse := response.Header.Get("Set-Cookie")
		Expect(cookieInResponse).ShouldNot(Equal(""))
		Expect(cookieInResponse).ShouldNot(Equal(cookie.String()))
	})
})
