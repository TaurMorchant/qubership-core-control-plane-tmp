package util

import (
	"context"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

func NewLoggerWrap(name string) *loggerWrapper {
	return WrapLogger(logging.GetLogger(name))
}

func WrapLogger(logger logging.Logger) *loggerWrapper {
	return &loggerWrapper{logger: logger}
}

type loggerWrapper struct {
	logger logging.Logger
}

func (l *loggerWrapper) DebugC(ctx context.Context, msg string, args ...any) {
	l.logger.DebugC(ctx, msg, args...)
}

func (l *loggerWrapper) InfoC(ctx context.Context, msg string, args ...any) {
	l.logger.InfoC(ctx, msg, args...)
}

func (l *loggerWrapper) WarnC(ctx context.Context, msg string, args ...any) {
	l.logger.WarnC(ctx, msg, args...)
}

// ErrorC logs provided error after the formatted message and returns the provided error.
//
// Example of usage:
//
//	if err != nil {
//	    return nil, log.ErrorC(ctx, err, "Could not do %s", "the thing")
//	}
//
// Output:
//
// [2023-02-02 15:01:10.502] [ERROR] [caller=the-logger] Could not do the thing:
//
// test: something bad happened
func (l *loggerWrapper) ErrorC(ctx context.Context, err error, msg string, args ...any) error {
	args = append(args, err)
	l.logger.ErrorC(ctx, msg+":\n %v", args...)
	return err
}
