// This package is used to initialize the application. It has dependencies on most
// other packages. Other packages can depend on it as a quick way to get access to
// all the dependencies.
package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/conf"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

type App struct {
	Config        *conf.App
	DB            *gorm.DB
	Repo          *Repos
	NeonClient    *neonapi.Client
	Register      *bgjobs.Register
	ProjectLocker *bgjobs.ProjectLocker
	RegionFilters []repos.Filter
}

func NewAppFromEnv() (*App, error) {
	cfg, err := conf.ParseEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	regionFilters := []repos.Filter{
		repos.FilterByRegionProvider(cfg.Provider),
	}
	if cfg.RegionFilters != "" {
		regionFilters = append(regionFilters, repos.RawFilter(cfg.RegionFilters))
	}
	log.Info(context.Background(), "using region filters", zap.Any("filters", regionFilters))

	db, err := connectDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	repo, err := createRepos(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create repos: %w", err)
	}

	neonClient := neonapi.NewClient(cfg.Provider, cfg.NeonAPIKey)
	register := bgjobs.NewRegister()
	projectLocker := bgjobs.NewProjectLocker()

	return &App{
		Config:        cfg,
		DB:            db,
		Repo:          repo,
		NeonClient:    neonClient,
		Register:      register,
		ProjectLocker: projectLocker,
		RegionFilters: regionFilters,
	}, nil
}

var (
	AlwaysOnQueryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "neonlight_alwayson_query_seconds",
		Help: "Time spent on each query",
	}, []string{"region", "driver"})
)

func (a *App) StartPrometheus() {
	go func() {
		mux := http.NewServeMux()
		prometheus.Register(AlwaysOnQueryTime)
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
	Query              *repos.QueryRepo
	SeqExitnodeProject *repos.Sequence
}

func createRepos(db *gorm.DB, cfg *conf.App) (*Repos, error) {
	err := db.AutoMigrate(
		&models.Region{},
		&models.Project{},
		&models.Sequence{},
		&models.GlobalRule{},
		&models.Query{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	if cfg.DebugDB {
		db = db.Debug()
	}

	regionRepo := repos.NewRegionRepo(db)
	projectRepo := repos.NewProjectRepo(db)
	sequenceRepo := repos.NewSequenceRepo(db)
	globalRuleRepo := repos.NewGlobalRuleRepo(db)
	queryRepo := repos.NewQueryRepo(db)

	exitnodeSeq, err := sequenceRepo.Get(fmt.Sprintf("exitnode-%s-project", cfg.Exitnode))
	if err != nil {
		return nil, fmt.Errorf("failed to get exitnode sequence: %w", err)
	}

	return &Repos{
		Region:             regionRepo,
		Project:            projectRepo,
		Sequence:           sequenceRepo,
		GlobalRule:         globalRuleRepo,
		Query:              queryRepo,
		SeqExitnodeProject: exitnodeSeq,
	}, nil
}
