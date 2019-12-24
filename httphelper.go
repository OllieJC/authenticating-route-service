package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func RedirectResponse(resp *http.Response, status int, url string) {
	body := fmt.Sprintf(`<head>
                         <meta http-equiv="refresh" content="0; URL=%s" />
                       </head>`, url)

	resp.StatusCode = status
	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	resp.Header.Add("Location", url)
}

func HTTPErrorResponse(err error) *http.Response {
	return &http.Response{
		Status:     "Error",
		StatusCode: 500,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(err.Error()))),
	}
}

func HTTPNotFoundResponse(err error) *http.Response {
	return &http.Response{
		Status:     "Not Found",
		StatusCode: 404,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("Not found"))),
	}
}
