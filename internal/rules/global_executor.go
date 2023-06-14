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

type GlobalExecutor struct {
	base *app.App
}

func NewGlobalExecutor(base *app.App) *GlobalExecutor {
	return &GlobalExecutor{base: base}
}

func (e *GlobalExecutor) ParseJSON(data json.RawMessage) (*Rule, error) {
	var desc rdesc.Rule
	err := json.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}

	return e.CreateFromDesc(desc)
}

func (e *GlobalExecutor) CreateFromDesc(desc rdesc.Rule) (*Rule, error) {
	impl, err := loadImpl(e.base, desc)
	if err != nil {
		return nil, err
	}

	return newRule(desc, impl)
}

func (e *GlobalExecutor) Execute(ctx context.Context, r *Rule) error {
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

func (e *GlobalExecutor) executeOnce(ctx context.Context, r *Rule) error {
	ctx = log.Into(ctx, string(r.desc.Act))
	return r.impl.Execute(ctx)
}

func (e *GlobalExecutor) executePeriodic(ctx context.Context, r *Rule, period *Period) error {
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
