package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/rules"
	"go.uber.org/zap"
)

func main() {
	_ = log.DefaultGlobals()
	ctx := context.Background()

	base, err := app.NewAppFromEnv()
	if err != nil {
		log.Fatal(ctx, "failed to init app", zap.Error(err))
	}

	// TODO: find a way to load it from env
	const defaultRule = `{"act": "do_global_rules", "args": {}, "periodic": "random(5,35)"}`

	var mainRule json.RawMessage = []byte(defaultRule)
	if len(os.Args) > 1 {
		mainRule = []byte(os.Args[1])
	}
	log.Info(ctx, "starting main rule", zap.Any("rule", mainRule))

	globalExecutor := rules.NewExecutor(base)
	rule, err := globalExecutor.ParseJSON(mainRule)
	if err != nil {
		log.Fatal(ctx, "failed to parse rule", zap.Error(err))
	}

	err = globalExecutor.Execute(ctx, rule)
	if err != nil {
		log.Fatal(ctx, "failed to execute rule", zap.Error(err))
	}

	base.Register.WaitAll(ctx)
}
