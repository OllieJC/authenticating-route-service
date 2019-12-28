package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func EmptyHTTPResponse(request *http.Request) (response *http.Response) {
	return &http.Response{
		Status:        "Not Implemented",
		StatusCode:    501,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          nil,
		ContentLength: 0,
		Request:       request,
		Header:        make(http.Header, 0),
	}
}

func RedirectResponse(resp *http.Response, status int, url string) {
	body := fmt.Sprintf(`<head>
                         <meta http-equiv="refresh" content="0; URL=%s" />
                       </head>`, url)

	resp.Status = "Redirect"
	resp.StatusCode = status
	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	resp.Header.Add("Location", url)
}

func HTTPErrorResponse(err error) *http.Response {
	response := EmptyHTTPResponse(nil)
	response.Status = "Error"
	response.StatusCode = 500
	response.Body = ioutil.NopCloser(bytes.NewReader([]byte(err.Error())))
	return response
}

func HTTPNotFoundResponse(err error) *http.Response {
	response := EmptyHTTPResponse(nil)
	response.Status = "Not Found"
	response.StatusCode = 404
	response.Body = ioutil.NopCloser(bytes.NewReader([]byte("Not found")))
	return response
}
