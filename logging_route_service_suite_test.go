package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoggingRouteService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LoggingRouteService Suite")
}
