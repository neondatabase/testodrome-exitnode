package rules

import (
	"context"
	"encoding/json"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
	"go.uber.org/zap"
)

type ctxkey int

const (
	ctxkeyInsidePeriodic ctxkey = iota
)

type Executor struct {
	base *app.App
}

func NewExecutor(base *app.App) *Executor {
	return &Executor{base: base}
}

func (e *Executor) ParseJSON(data json.RawMessage) (*Rule, error) {
	var desc rdesc.Rule
	err := json.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}

	return e.CreateFromDesc(desc)
}

func (e *Executor) CreateFromDesc(desc rdesc.Rule) (*Rule, error) {
	impl, err := loadImpl(e.base, e, desc)
	if err != nil {
		return nil, err
	}

	return newRule(desc, impl)
}

func (e *Executor) Execute(ctx context.Context, r *Rule) error {
	var insidePeriodic bool
	if val, ok := ctx.Value(ctxkeyInsidePeriodic).(bool); ok {
		insidePeriodic = val
	}

	// can't execute nested periodic rules
	isPeriodic := r.period != nil && !insidePeriodic
	if isPeriodic {
		return e.executePeriodic(ctx, r, r.period)
	}
	return e.executeOnce(ctx, r)
}

func (e *Executor) executeOnce(ctx context.Context, r *Rule) error {
	ctx = log.Into(ctx, string(r.desc.Act))
	return r.impl.Execute(ctx)
}

func (e *Executor) executePeriodic(ctx context.Context, r *Rule, period *Period) error {
	ctx = context.WithValue(ctx, ctxkeyInsidePeriodic, true)
	ctx = log.Into(ctx, "periodic")

	// TODO: watch for ctx.Done()
	for {
		err := e.executeOnce(ctx, r)
		if err != nil {
			// TODO: add option to propagate errors
			log.Error(ctx, "rule execution failed", zap.Error(err))
		}

		period.Sleep(ctx)
	}
}
