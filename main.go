package main

import (
	"context"
	"encoding/json"
	"time"

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

	globalExecutor := rules.NewGlobalExecutor(base)
	rule, err := globalExecutor.ParseJSON(json.RawMessage(`{"act": "create_project", "args": {"interval": "10m"}}`))
	if err != nil {
		log.Fatal(ctx, "failed to parse rule", zap.Error(err))
	}

	err = globalExecutor.Execute(ctx, rule)
	if err != nil {
		log.Fatal(ctx, "failed to execute rule", zap.Error(err))
	}

	time.Sleep(time.Second * 20)

	// var ruleList []rules.ExecutableRule
	// // create new project every 10 minutes (in each region)
	// ruleList = append(ruleList, rules.NewCreateProject(base, time.Minute*10))
	// // delete projects if there are > 5 (in each region)
	// ruleList = append(ruleList, rules.NewDeleteProject(base, 5))
}
