package configurator

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// GoogleEmailDomain is a type which contains Google oauth settings
type GoogleEmailDomain struct {
	Domain            string `yaml:"domain"`
	OAuthClientID     string `yaml:"google_oauth_client_id"`
	OAuthClientSecret string `yaml:"google_oauth_client_secret"`
}

// DomainConfig is the type which an entire site's config is within
type DomainConfig struct {
	Domain             string              `yaml:"domain"`
	AuthPageTitle      string              `yaml:"auth_pages_title"`
	Debug              bool                `yaml:"debug"`
	Enabled            bool                `yaml:"enabled"`
	GoogleEmailDomains []GoogleEmailDomain `yaml:"google_email_domains"`
	SessionCookieName  string              `yaml:"session_cookie_name"`
	SessionServerToken string              `yaml:"session_server_token"`
	SecurityHeaders    map[string]string   `yaml:"security_headers"`
	SkipTLSValidation  bool                `yaml:"skip_tls_validation"`
}

// Config is the master configuration type, it has an array of DomainConfig objects
type Config struct {
	DomainConfigs []DomainConfig `yaml:"domains"`
}

// ReadConfigFile takes a yaml filename, attempts to parse and returns Config object
func ReadConfigFile(filename string) (Config, error) {
	var c Config

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

// GetDomainConfigFromRequest returns DomainConfig (and error) from a request and DOMAIN_CONFIG_FILEPATH env var
func GetDomainConfigFromRequest(request *http.Request) (DomainConfig, error) {
	return GetDomainConfig(request.URL.Hostname(), os.Getenv("DOMAIN_CONFIG_FILEPATH"))
}

// Get returns the DomainConfig for a specific domain
func (c Config) Get(domain string) DomainConfig {
	var dc DomainConfig
	for _, d := range c.DomainConfigs {
		if domain == d.Domain && d.Enabled {
			dc = d
			break
		}
	}
	return dc
}

// GetGoogleEmailDomain returns the GoogleEmailDomain for a specific domain out of DomainConfig
func (c DomainConfig) GetGoogleEmailDomain(domain string) GoogleEmailDomain {
	var ged GoogleEmailDomain
	for _, d := range c.GoogleEmailDomains {
		if domain == d.Domain {
			ged = d
			break
		}
	}
	return ged
}

// GetDomainConfig returns DomainConfig (and error) from domain and filepath
func GetDomainConfig(domain string, filename string) (DomainConfig, error) {
	var err error

	c, err := ReadConfigFile(filename)
	if err != nil {
		return DomainConfig{}, err
	}

	dc := c.Get(domain)
	if dc.Enabled == false {
		err = errors.New("Domain not found in config file")
		return DomainConfig{}, err
	}

	return dc, nil
}
