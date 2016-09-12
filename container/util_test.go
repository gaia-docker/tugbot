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

func TestInFilterOrList_ListTrue(t *testing.T) {
	assert.True(t, inFilterOrList("delete", "resize, start, untag, delete"))
}

func TestInFilterOrList_ListFalse(t *testing.T) {
	assert.False(t, inFilterOrList("what?", "resize, start, untag, delete"))
}

func TestInFilterOrList_FilterTrue(t *testing.T) {
	assert.True(t, inFilterOrList("cont123", "re2:^cont"))
}

func TestInFilterOrList_FilterFalse(t *testing.T) {
	assert.False(t, inFilterOrList("cont123", "re2:^NO"))
}

func TestMapContains_True(t *testing.T) {
	m := map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}
	assert.True(t, mapContains(m, []string{"k2", "v2"}))
}

func TestMapContains_SingleTrue(t *testing.T) {
	m := map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}
	assert.True(t, mapContains(m, []string{"k2"}))
}

func TestMapContains_False(t *testing.T) {
	m := map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}
	assert.False(t, mapContains(m, []string{"x", "y"}))
}
