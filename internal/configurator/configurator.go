package configurator

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// LoginEmailDomain is a type which contains Google oauth settings
type LoginEmailDomain struct {
	Domain   string `yaml:"domain"`
	Provider string `yaml:"provider"`
}

// DomainConfig is the type which an entire site's config is within
type DomainConfig struct {
	Domain                  string             `yaml:"domain"`
	AuthPageTitle           string             `yaml:"auth_pages_title"`
	Enabled                 bool               `yaml:"enabled"`
	GoogleOAuthClientID     string             `yaml:"google_oauth_client_id"`
	GoogleOAuthClientSecret string             `yaml:"google_oauth_client_secret"`
	LoginEmailDomains       []LoginEmailDomain `yaml:"login_email_domains"`
	SessionCookieName       string             `yaml:"session_cookie_name"`
	SessionServerToken      string             `yaml:"session_server_token"`
	SecurityHeaders         map[string]string  `yaml:"security_headers"`
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
		//fmt.Printf("ReadConfigFile: Couldn't read: '%s'\n", filename)
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
	hsn := request.URL.Hostname()
	dcf := os.Getenv("DOMAIN_CONFIG_FILEPATH")
	if dcf == "" {
		dcf = "config/default.yml"
	}
	//fmt.Printf("hsn: %s\n", hsn)
	//fmt.Printf("dcf: %s\n", dcf)
	return GetDomainConfig(hsn, dcf)
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

// GetLoginEmailDomain returns the GoogleEmailDomain for a specific domain out of DomainConfig
func (c DomainConfig) GetLoginEmailDomain(domain string) LoginEmailDomain {
	var led LoginEmailDomain
	for _, d := range c.LoginEmailDomains {
		if domain == d.Domain {
			led = d
			break
		}
	}
	return led
}

// GetDomainConfig returns DomainConfig (and error) from domain and filepath
func GetDomainConfig(domain string, filename string) (DomainConfig, error) {
	var err error

	c, err := ReadConfigFile(filename)
	if err != nil {
		//fmt.Printf("GetDomainConfig: Couldn't find '%s' in '%s'\n", domain, filename)
		return DomainConfig{}, err
	}

	dc := c.Get(domain)
	if dc.Enabled == false {
		err = errors.New("Domain not found in config file")
		return DomainConfig{}, err
	}

	return dc, nil
}
