package internal_test

import (
	s "authenticating-route-service/internal"
	h "authenticating-route-service/internal/httphelper"
	"os"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Internal", func() {
	// Defaults to test when the tests run
	s.StaticAssetPath = "../web/static"

	h.TemplatePath = "../web/template"

	os.Setenv("DOMAIN_CONFIG_FILEPATH", "../test/data/example.yml")
})
