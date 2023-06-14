// This package is used to initialize the application. It has dependencies on most
// other packages. Other packages can depend on it as a quick way to get access to
// all the dependencies.
package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/conf"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	Config     *conf.App
	DB         *gorm.DB
	Repo       *Repos
	NeonClient *neonapi.Client
	Register   *bgjobs.Register
}

func NewAppFromEnv() (*App, error) {
	cfg, err := conf.ParseEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	db, err := connectDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	repos, err := createRepos(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create repos: %w", err)
	}

	neonClient := neonapi.NewClient(cfg.Provider, cfg.NeonApiKey)
	register := bgjobs.NewRegister()

	return &App{
		Config:     cfg,
		DB:         db,
		Repo:       repos,
		NeonClient: neonClient,
		Register:   register,
	}, nil
}

func (a *App) StartPrometheus() {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(a.Config.PrometheusBind, mux)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(context.TODO(), "prometheus server error", zap.Error(err))
		}
	}()
}

func connectDB(cfg *conf.App) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

type Repos struct {
	Region             *repos.RegionRepo
	Project            *repos.ProjectRepo
	Sequence           *repos.SequenceRepo
	GlobalRule         *repos.GlobalRuleRepo
	SeqExitnodeProject *repos.Sequence
}

func createRepos(db *gorm.DB, cfg *conf.App) (*Repos, error) {
	err := db.AutoMigrate(
		&models.Region{},
		&models.Project{},
		&models.Sequence{},
		&models.GlobalRule{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	if cfg.DbDebug {
		db = db.Debug()
	}

	regionRepo := repos.NewRegionRepo(db)
	projectRepo := repos.NewProjectRepo(db)
	sequenceRepo := repos.NewSequenceRepo(db)
	globalRuleRepo := repos.NewGlobalRuleRepo(db)

	exitnodeSeq, err := sequenceRepo.Get(fmt.Sprintf("exitnode-%s-project", cfg.Exitnode))
	if err != nil {
		return nil, fmt.Errorf("failed to get exitnode sequence: %w", err)
	}

	return &Repos{
		Region:             regionRepo,
		Project:            projectRepo,
		Sequence:           sequenceRepo,
		GlobalRule:         globalRuleRepo,
		SeqExitnodeProject: exitnodeSeq,
	}, nil
}
