package rules

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/rdesc"
)

// Rule is a fully initialized rule that can be executed via global executor.
type Rule struct {
	desc    rdesc.Rule
	impl    RuleImpl
	period  *Period
	lastRun *time.Time
}

func newRule(desc rdesc.Rule, impl RuleImpl) (*Rule, error) {
	period, err := parsePeriod(desc.Periodic)
	if err != nil {
		return nil, err
	}

	return &Rule{
		desc:   desc,
		impl:   impl,
		period: period,
	}, nil
}

type Period struct {
	min uint
	max uint
}

func (p *Period) Sleep(ctx context.Context) {
	val := p.min + uint(rand.Intn(int(p.max-p.min+1)))
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(val) * time.Second):
	}
}

func parsePeriod(str string) (*Period, error) {
	if str == "" {
		return nil, nil
	}

	var min, max uint

	_, err := fmt.Sscanf(str, "random(%d,%d)", &min, &max)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period: %w", err)
	}

	if min > max {
		return nil, fmt.Errorf("min(%d) > max(%d)", min, max)
	}

	return &Period{
		min: min,
		max: max,
	}, nil
}
