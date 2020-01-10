package configurator_test

import (
	"fmt"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service/internal/configurator"
)

var _ = Describe("configurator", func() {

	It("should return sensible objects from ReadConfigFile", func() {
		c, err := s.ReadConfigFile("../../test/data/example.yml")
		Expect(err).ToNot(HaveOccurred())

		// example.yml has two entries
		Expect(len(c.DomainConfigs)).To(BeEquivalentTo(2))

		// first item domain
		Expect(c.DomainConfigs[0].Domain).To(Equal("example.local"))

		// if no value set, should default
		Expect(c.DomainConfigs[1].Enabled).To(Equal(false))

		Expect(c.DomainConfigs[0].LoginEmailDomains[0].Provider).To(Equal("google"))

		printme := false
		if printme {
			for _, d := range c.DomainConfigs {
				fmt.Println("------------")
				fmt.Println(d.Domain)
				fmt.Println(d.AuthPageTitle)
				fmt.Println(d.Enabled)
				for i, g := range d.LoginEmailDomains {
					fmt.Printf("g:%d: %s\n", i, g.Domain)
					fmt.Printf("g:%d: %s\n", i, g.Provider)
					fmt.Printf("g:%d: %s\n", i, g.OAuthClientID)
					fmt.Printf("g:%d: %s\n", i, g.OAuthClientSecret)
				}
				fmt.Println(d.SessionCookieName)
				fmt.Println(d.SessionServerToken)

				for k, s := range d.SecurityHeaders {
					fmt.Printf("d: %s: %s\n", k, s)
				}

				for i, s := range d.UnauthenticatedPaths {
					fmt.Printf("d: %d: %s\n", i, s)
				}
			}
		}
	})

	It("should return an error from GetDomainConfigFromRequest with no domain", func() {
		request, _ := http.NewRequest("GET", "http://not-valid.local", nil)

		_, err := s.GetDomainConfigFromRequest(request)

		Expect(err).To(HaveOccurred())
	})

	It("should return an object from GetDomainConfigFromRequest", func() {
		os.Setenv("DOMAIN_CONFIG_FILEPATH", "../../test/data/example.yml")
		request, _ := http.NewRequest("GET", "http://example.local", nil)

		dc, err := s.GetDomainConfigFromRequest(request)

		Expect(err).ToNot(HaveOccurred())
		Expect(dc).Should(BeAssignableToTypeOf(s.DomainConfig{}))
		Expect(dc.Domain).Should(Equal("example.local"))
	})

	It("should not return an object from GetDomainConfig if enabled is not set", func() {
		var err error

		_, err = s.GetDomainConfig("testing.uk", "../../test/data/example.yml")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Domain not found in config file"))
	})

	It("should not return an object from GetDomainConfig if filename is empty", func() {
		_, err := s.GetDomainConfig("example.local", "")

		Expect(err).To(HaveOccurred())
	})

	It("should not return an object from GetDomainConfig if file bad", func() {
		_, err := s.GetDomainConfig("example.local", "../../test/data/bad.yml")

		Expect(err).To(HaveOccurred())
	})

	It("should return LoginEmailDomain from DomainConfig", func() {
		dc, err := s.GetDomainConfig("example.local", "../../test/data/example.yml")
		Expect(err).ToNot(HaveOccurred())

		ged := dc.GetLoginEmailDomain("second.example.local", "None")
		Expect(ged.Provider).To(Equal("none"))
	})

	It("should return correct LoginEmailDomain from DomainConfig", func() {
		dc, err := s.GetDomainConfig("example.local", "../../test/data/example.yml")
		Expect(err).ToNot(HaveOccurred())

		ged := dc.GetLoginEmailDomain("third.example.local", "GitHub")
		Expect(ged.Provider).To(Equal("github"))
	})

	It("should return multiple providers with GetLoginProviders", func() {
		dc, err := s.GetDomainConfig("example.local", "../../test/data/example.yml")
		Expect(err).ToNot(HaveOccurred())

		providers := dc.GetLoginProviders("third.example.local")
		Expect(len(providers)).Should(BeNumerically("==", 2))
	})

	It("should return true when visitng /test/unauth with example.yml", func() {
		request, _ := http.NewRequest("GET", "http://example.local/test/unauth", nil)

		res := s.IsUnauthPath(request)

		Expect(res).To(BeTrue())
	})

	It("should return false when visitng /test/auth with example.yml", func() {
		request, _ := http.NewRequest("GET", "http://example.local/test/auth", nil)

		res := s.IsUnauthPath(request)

		Expect(res).To(BeFalse())
	})

	It("should return false when visitng bad host", func() {
		request, _ := http.NewRequest("GET", "http://not-valid.local/test/unauth", nil)

		res := s.IsUnauthPath(request)

		Expect(res).To(BeFalse())
	})
})
