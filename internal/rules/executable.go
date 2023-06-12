package rules

import "context"

// Just some rule that can be executed periodically.
type ExecutableRule interface {
	Execute(ctx context.Context) error
}
