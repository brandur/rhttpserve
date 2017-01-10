package common

import (
	"fmt"
	"os"
	"strings"
)

// ExitWithError exits the program after printing the given error's message.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// Message generates a message payload based off a path and expiry time.
func Message(path string, expiresAt int64) []byte {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return []byte(fmt.Sprintf("%v|%v", path, expiresAt))
}
