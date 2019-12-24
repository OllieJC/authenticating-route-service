package main_test

import (
	"errors"
	"io/ioutil"
	"net/http"

	//"github.com/jarcoal/httpmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s "authenticating-route-service"
)

var _ = Describe("HTTPHelper", func() {
	It("should return an error page with HTTPErrorResponse", func() {
		const errStr = "Test error."
		testErr := errors.New(errStr)

		var response *http.Response
		response = s.HTTPErrorResponse(testErr)

		Expect(response.StatusCode).To(Equal(500))

		bodyBytes, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(bodyBytes)).To(Equal(errStr))
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
})
