package common

import (
	"fmt"
	"os"
)

// ExitWithError exits the program after printing the given error's message.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// Message generates a message payload based off a path and expiry time.
func Message(remote, path string, expiresAt int64) []byte {
	return []byte(fmt.Sprintf("%v|%v|%v", remote, path, expiresAt))
}
