package internal_test

import (
	s "authenticating-route-service/internal"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Internal", func() {
	// Defaults to test when the tests run
	s.TemplatePath = "../web/template"
	s.StaticAssetPath = "../web/static"
	s.InternalTest = true
})
