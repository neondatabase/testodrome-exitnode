package log

import (
	"context"

	"go.uber.org/zap"
)

type ctxkey string

const (
	loggerContextKey ctxkey = "logger"
)

func createGlobalLogger() (*zap.Logger, error) {
	return zap.NewDevelopment(zap.AddCallerSkip(1))
}

// DefaultGlobals replaces global zap logger with custom default configuration.
func DefaultGlobals() func() {
	return zap.ReplaceGlobals(zap.Must(createGlobalLogger()))
}

// FromContext returns logger from context if set. Otherwise returns global logger.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.L()
	}
	if logger, ok := ctx.Value(loggerContextKey).(*zap.Logger); ok {
		return logger
	}
	return zap.L()
}

// Into returns context with a new span.
func Into(ctx context.Context, span string) context.Context {
	var logger = FromContext(ctx).WithOptions().Named(span)
	return context.WithValue(ctx, loggerContextKey, logger)
}

// With appends fields to logger in context.
func With(ctx context.Context, args ...zap.Field) context.Context {
	var logger = FromContext(ctx).With(args...)
	return context.WithValue(ctx, loggerContextKey, logger)
}

func Debug(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Debug(msg, args...)
}

func Info(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Error(msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Fatal(msg, args...)
}
