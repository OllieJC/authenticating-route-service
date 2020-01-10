package google_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDebugprint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Google Suite")
}
