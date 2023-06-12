// This package is used to initialize the application. It has dependencies on most
// other packages. Other packages can depend on it as a quick way to get access to
// all the dependencies.
package app

import (
	"fmt"
	"net/http"

	"github.com/petuhovskiy/neon-lights/internal/conf"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	Config     *conf.App
	DB         *gorm.DB
	Repo       *Repos
	NeonClient *neonapi.Client
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

	return &App{
		Config:     cfg,
		DB:         db,
		Repo:       repos,
		NeonClient: neonClient,
	}, nil
}

func (a *App) StartPrometheus() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(a.Config.PrometheusBind, mux)
	if err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatal("prometheus server error")
	}
}

func connectDB(cfg *conf.App) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db = db.Debug()
	return db, nil
}

type Repos struct {
	Region             *repos.RegionRepo
	Project            *repos.ProjectRepo
	Sequence           *repos.SequenceRepo
	SeqExitnodeProject *repos.Sequence
}

func createRepos(db *gorm.DB, cfg *conf.App) (*Repos, error) {
	err := db.AutoMigrate(
		&models.Region{},
		&models.Project{},
		&models.Sequence{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	regionRepo := repos.NewRegionRepo(db)
	projectRepo := repos.NewProjectRepo(db)
	sequenceRepo := repos.NewSequenceRepo(db)

	exitnodeSeq, err := sequenceRepo.Get(fmt.Sprintf("exitnode-%s-project", cfg.Exitnode))
	if err != nil {
		return nil, fmt.Errorf("failed to get exitnode sequence: %w", err)
	}

	return &Repos{
		Region:             regionRepo,
		Project:            projectRepo,
		Sequence:           sequenceRepo,
		SeqExitnodeProject: exitnodeSeq,
	}, nil
}
