package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	"github.com/petuhovskiy/neon-lights/internal/rules"
	"github.com/petuhovskiy/neon-lights/pkg/conf"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	cfg, err := conf.ParseEnv()
	if err != nil {
		log.WithError(err).Fatal("failed to parse config from env")
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.PrometheusBind, mux)
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("prometheus server error")
		}
	}()

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}
	db = db.Debug()

	db.AutoMigrate(
		&models.Region{},
		&models.Project{},
		&models.Sequence{},
	)

	regionRepo := repos.NewRegionRepo(db)
	projectRepo := repos.NewProjectRepo(db)
	sequenceRepo := repos.NewSequenceRepo(db)

	exitnodeSeq, err := sequenceRepo.Get(fmt.Sprintf("exitnode-%s-project", cfg.Exitnode))
	if err != nil {
		log.WithError(err).Fatal("failed to get exitnode sequence")
	}

	neonClient := neonapi.NewClient(cfg.Provider, cfg.NeonApiKey)

	var ruleList []rules.ExecutableRule
	ruleList = append(ruleList, &rules.CreateProject{
		GapDuration: time.Minute * 10,
		Provider:    cfg.Provider,
		RegionRepo:  regionRepo,
		ProjectRepo: projectRepo,
		Sequence:    exitnodeSeq,
		NeonClient:  neonClient,
		Config:      cfg,
	})

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
