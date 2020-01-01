package debugprint_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "authenticating-route-service/pkg/debugprint"
)

var _ = Describe("debugprint", func() {
	const testString = "Testing123."

	It("should do nothing when env var is not set", func() {
		os.Unsetenv("DEBUG")

		t := Debugfln("Test: %s", testString)

		Expect(t).ToNot(ContainSubstring(testString))
		Expect(t).To(Equal(""))
	})

	It("should do nothing when env var is set to something that isn't a bool", func() {
		os.Setenv("DEBUG", "abc123")

		t := Debugfln("Test: %s", testString)

		Expect(t).ToNot(ContainSubstring(testString))
		Expect(t).To(Equal(""))
	})

	It("should print to stdout when env var is set", func() {
		os.Setenv("DEBUG", "true")

		t := Debugfln("Test: %s", testString)

		Expect(t).To(ContainSubstring(testString))
		Expect(t).ToNot(HaveSuffix("\n"))
	})
})
