package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUrlValidator_IsValidPath(t *testing.T) {
	assert.False(t, DefaultUrlValidator.IsValidPath(""))
	assert.False(t, DefaultUrlValidator.IsValidPath("/pathFrom /"))
	assert.True(t, DefaultUrlValidator.IsValidPath("/api"))
}
