package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage(t *testing.T) {
	assert.Equal(t, "remote|path/to/file|123", string(Message("remote", "path/to/file", 123)))
}
