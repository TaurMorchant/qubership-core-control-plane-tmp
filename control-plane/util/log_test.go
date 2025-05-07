package util

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoggerWrapper(t *testing.T) {
	ctx := context.Background()
	log := NewLoggerWrap("the-logger")
	log.InfoC(ctx, "This is just number one: %d", 1)
	log.WarnC(ctx, "%s - I warned you", "Mesh is complex")
	err := errors.New("test: something bad happened")
	err2 := log.ErrorC(ctx, err, "Could not do %s", "the thing")
	assert.Equal(t, err, err2)
}
