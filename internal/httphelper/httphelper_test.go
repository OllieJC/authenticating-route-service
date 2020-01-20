package httphelper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal/httphelper"
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

var _ = Describe("HTTPHelper", func() {
	os.Setenv("DOMAIN_CONFIG_FILEPATH", "../../test/data/example.yml")
	s.TemplatePath = "../../web/template"
	redCookieName := "_redirectPath"

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

	It("should set expiry in past with RemoveCookie", func() {
		response := s.EmptyHTTPResponse(nil)

		request, _ := http.NewRequest("GET", "http://example.local", nil)

		s.RemoveCookie(request, response, "_session")

		cookieRawVal := response.Header.Get("Set-Cookie")
		Expect(cookieRawVal).ToNot(BeNil())

		t, err := parseCookieTime(cookieRawVal)
		Expect(err).NotTo(HaveOccurred())

		// cookie time should not equal the default time
		Expect(t.Unix()).ShouldNot(BeEquivalentTo(-62135596800))

		// cookie time should be less than current time
		Expect(t.Unix()).Should(BeNumerically("<", time.Now().Unix()))
	})

	It("should return default path with no cookie set and RedirectCookiePath", func() {
		response := s.EmptyHTTPResponse(nil)
		request, _ := http.NewRequest("GET", "http://example.local", nil)

		resPath := s.RedirectCookieURI(request, response, redCookieName)
		Expect(resPath).To(Equal("/"))

		sc := response.Header["Set-Cookie"]
		Expect(sc).To(BeNil())
	})

	It("should return default path with no cookie set and RedirectCookiePath", func() {
		response := s.EmptyHTTPResponse(nil)
		request, _ := http.NewRequest("GET", "http://example.local", nil)

		cookie := &http.Cookie{Name: redCookieName, Value: "http://test.com/testing123/abc"}
		request.AddCookie(cookie)

		resPath := s.RedirectCookieURI(request, response, redCookieName)
		Expect(resPath).To(Equal("/testing123/abc"))

		sc := response.Header["Set-Cookie"]
		Expect(sc).ToNot(BeNil())
		Expect(sc[0]).To(ContainSubstring(redCookieName))
	})

	It("should return default path with no cookie set and RedirectCookiePath", func() {
		response := s.EmptyHTTPResponse(nil)
		request, _ := http.NewRequest("GET", "http://example.local", nil)

		cookie := &http.Cookie{Name: redCookieName, Value: "http://test.com/testing123/abc?test=123&abc=456"}
		request.AddCookie(cookie)

		resPath := s.RedirectCookieURI(request, response, redCookieName)
		Expect(resPath).To(Equal("/testing123/abc?test=123&abc=456"))

		sc := response.Header["Set-Cookie"]
		Expect(sc).ToNot(BeNil())
		Expect(sc[0]).To(ContainSubstring(redCookieName))
	})

	It("should return default path with no cookie set and RedirectCookiePath", func() {
		response := s.EmptyHTTPResponse(nil)
		request, _ := http.NewRequest("GET", "http://example.local", nil)

		cookie := &http.Cookie{Name: redCookieName, Value: "/testing123/def"}
		request.AddCookie(cookie)

		resPath := s.RedirectCookieURI(request, response, redCookieName)
		Expect(resPath).To(Equal("/testing123/def"))

		sc := response.Header["Set-Cookie"]
		Expect(sc).ToNot(BeNil())
		Expect(sc[0]).To(ContainSubstring(redCookieName))
	})
})
