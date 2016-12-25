package common

import (
	"fmt"
	"os"
	"strings"
)

func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func Message(path string, expiresAt int64) []byte {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return []byte(fmt.Sprintf("%v|%v", path, expiresAt))
}
