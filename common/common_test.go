package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage(t *testing.T) {
	assert.Equal(t, "/path/to/file|123", string(Message("path/to/file", 123)))
}
