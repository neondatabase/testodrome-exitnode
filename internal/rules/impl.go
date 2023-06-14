package rules

import (
	"context"
	"fmt"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
)

var ErrUnknownRule = fmt.Errorf("unknown rule")

// One of the rule implementations.
type RuleImpl interface {
	Execute(ctx context.Context) error
}

func loadImpl(base *app.App, desc rdesc.Rule) (RuleImpl, error) {
	switch desc.Act {
	case rdesc.ActCreateProject:
		return NewCreateProject(base, desc.Args)
	case rdesc.ActDeleteProject:
		return NewDeleteProject(base, desc.Args)
	default:
		return nil, fmt.Errorf("unknown rule act %s: %w", desc.Act, ErrUnknownRule)
	}
}
