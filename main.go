package main

import (
	"context"
	"math/rand"
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

	var ruleList []rules.ExecutableRule
	// create new project every 10 minutes (in each region)
	ruleList = append(ruleList, rules.NewCreateProject(base, time.Minute*10))
	// delete projects if there are > 5 (in each region)
	ruleList = append(ruleList, rules.NewDeleteProject(base, 5))

	for {
		for _, rule := range ruleList {
			err := rule.Execute(ctx)
			if err != nil {
				log.Error(ctx, "rule execution failed", zap.Error(err))
			}
		}

		delay := rand.Intn(30) + 5
		time.Sleep(time.Second * time.Duration(delay))
	}
}
