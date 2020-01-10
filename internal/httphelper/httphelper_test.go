package httphelper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal/httphelper"
)

var _ = Describe("HTTPHelper", func() {
	os.Setenv("DOMAIN_CONFIG_FILEPATH", "../../test/data/example.yml")
	s.TemplatePath = "../../web/template"

	It("should return an error page with HTTPErrorResponse", func() {
		const errStr = "Test error."
		testErr := errors.New(errStr)

		var response *http.Response
		response = s.HTTPErrorResponse(testErr)

		Expect(response.StatusCode).To(Equal(500))

		bodyBytes, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).To(ContainSubstring(errStr))
	})

	It("should return an 404 page with HTTPErrorResponse", func() {
		const errStr = "Test error."
		testErr := errors.New(errStr)

		var response *http.Response
		response = s.HTTPNotFoundResponse(testErr)

		Expect(response.StatusCode).To(Equal(404))

		bodyBytes, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).ToNot(ContainSubstring(errStr))
	})

	It("should add sensible security header defaults with AddSecurityHeaders", func() {

		request := httptest.NewRequest("GET", "http://testing.uk/", nil)
		response := s.EmptyHTTPResponse(nil)

		s.AddSecurityHeaders(request, response)

		xxp := response.Header.Get("X-Xss-Protection")
		Expect(xxp).To(Equal("1; mode=block"))

		xcto := response.Header.Get("X-Content-Type-Options")
		Expect(xcto).To(Equal("nosniff"))

		xfo := response.Header.Get("X-Frame-Options")
		Expect(xfo).To(Equal("DENY"))

		csp := response.Header.Get("Content-Security-Policy")
		Expect(csp).To(Equal("default-src 'self'"))

		rp := response.Header.Get("Referrer-Policy")
		Expect(rp).To(Equal("strict-origin-when-cross-origin"))

		fp := response.Header.Get("Feature-Policy")
		Expect(fp).To(Equal("vibrate 'none'; geolocation 'none'; microphone 'none'; camera 'none'; payment 'none'; notifications 'none';"))
	})

	It("should allow headers to be overrode in AddSecurityHeaders", func() {

		request := httptest.NewRequest("GET", "http://example.local/", nil)
		response := s.EmptyHTTPResponse(nil)
		s.AddSecurityHeaders(request, response)

		_, xxp := response.Header["Referrer-Policy"]
		Expect(xxp).To(Equal(false))

		xcto := response.Header.Get("X-Content-Type-Options")
		Expect(xcto).To(Equal("nosniff"))

		xfo := response.Header.Get("X-Frame-Options")
		Expect(xfo).To(Equal("ALLOW"))
	})
})
