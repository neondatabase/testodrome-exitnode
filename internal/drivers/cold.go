package drivers

import "context"

type ctxKey string

const (
	// This flag is set when we know that we had several queries in a row and
	// current query is not the first one.
	notColdKey ctxKey = "not_cold"
)

// NotColdContext returns a new context with the not cold flag set.
func NotColdContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, notColdKey, true)
}

// IsNotCold returns true if the context has the not cold flag set.
func IsNotCold(ctx context.Context) bool {
	return ctx.Value(notColdKey) == true
}
