package common

import (
	"fmt"
	"strings"
)

func Message(path string, expiresAt int64) []byte {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return []byte(fmt.Sprintf("%v|%v", path, expiresAt))
}
