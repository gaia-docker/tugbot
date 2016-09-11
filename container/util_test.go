package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContains_True(t *testing.T) {
	s := []string{"resize", "start", "untag", "delete"}

	result := sliceContains("delete", s)

	assert.True(t, result)
}

func TestSliceContains_False(t *testing.T) {
	s := []string{"resize", "start", "untag", "delete"}

	result := sliceContains("foo", s)

	assert.False(t, result)
}
