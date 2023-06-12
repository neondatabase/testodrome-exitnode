package main

import (
	"math/rand"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/rules"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	base, err := app.NewAppFromEnv()
	if err != nil {
		log.WithError(err).Fatal("failed to init app")
	}

	var ruleList []rules.ExecutableRule
	// create new project every 10 minutes (in each region)
	ruleList = append(ruleList, rules.NewCreateProject(base, time.Minute*10))
	// delete projects if there are > 5 (in each region)
	ruleList = append(ruleList, rules.NewDeleteProject(base, 5))

	for {
		for _, rule := range ruleList {
			err := rule.Execute()
			if err != nil {
				log.WithError(err).Error("rule execution error")
			}
		}

		delay := rand.Intn(30) + 5
		time.Sleep(time.Second * time.Duration(delay))
	}
}
