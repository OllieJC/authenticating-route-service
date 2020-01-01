package debugprint

import (
	"fmt"
	"os"
	"strconv"
)

// Debugfln prints a new line to stdout if "DEBUG" env var is set to "true"
func Debugfln(fStr string, args ...interface{}) string {
	if DebugOption() {
		res := fmt.Sprintf(fStr, args...)
		fmt.Fprintln(os.Stdout, res)
		return res
	}

	return ""
}

// DebugOption returns true if "DEBUG" env var is set to "true"
func DebugOption() bool {
	res := os.Getenv("DEBUG")

	if len(res) == 0 {
		res = "false"
	}

	rb, _ := strconv.ParseBool(res)
	return rb
}
