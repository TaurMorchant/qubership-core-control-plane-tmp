package cookie

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGenerateCookieName(t *testing.T) {
	cookie1 := NameGenerator.GenerateCookieName("sticky")
	cookie2 := NameGenerator.GenerateCookieName("sticky")
	cookie3 := NameGenerator.GenerateCookieName("sticky")
	assert.True(t, strings.HasPrefix(cookie1, "sticky-"))
	assert.True(t, strings.HasPrefix(cookie2, "sticky-"))
	assert.True(t, strings.HasPrefix(cookie3, "sticky-"))
	assert.NotEqual(t, cookie1, cookie2)
	assert.NotEqual(t, cookie1, cookie3)
	assert.NotEqual(t, cookie2, cookie3)
}
