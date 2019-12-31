package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

func debug(fStr string, args ...interface{}) {
	var DebugOut io.Writer = ioutil.Discard
	if debugOption() {
		DebugOut = os.Stdout
	}
	fmt.Fprintf(DebugOut, fStr+"\n", args...)
}

var _debugOption string = ""

func debugOption() bool {
	if _debugOption == "" {
		dStr := os.Getenv("DEBUG")
		if len(dStr) != 0 {
			_debugOption = dStr
		} else {
			_debugOption = "false"
		}
	}

	res, _ := strconv.ParseBool(_debugOption)
	return res
}
