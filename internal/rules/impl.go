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

func loadImpl(base *app.App, executor *Executor, desc rdesc.Rule) (RuleImpl, error) {
	switch desc.Act {
	case rdesc.ActCreateProject:
		return NewCreateProject(base, desc.Args)
	case rdesc.ActDeleteProject:
		return NewDeleteProject(base, desc.Args)
	case rdesc.ActDoGlobalRules:
		return NewDoGlobalRules(base, executor, desc.Args)
	case rdesc.ActQueryProject:
		return NewQueryProject(base, desc.Args)
	case rdesc.ActTest:
		return NewTestRule(base, desc.Args)
	default:
		return nil, fmt.Errorf("unknown rule act %s: %w", desc.Act, ErrUnknownRule)
	}
}
