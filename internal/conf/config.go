package conf

import (
	"github.com/caarlos0/env/v6"
)

type App struct {
	PrometheusBind string `env:"PROMETHEUS_BIND" envDefault:":2112"`

	// PostgresDSN is a DSN for the postgres.
	PostgresDSN string `env:"POSTGRES_DSN,required"`

	// Exitnode is a name of the current node.
	Exitnode string `env:"EXITNODE" envDefault:"local-laptop"`

	// Provider is a name/domain of the current provider.
	Provider string `env:"PROVIDER" envDefault:"stage.neon.tech"`

	// NeonAPIKey is an API key for the neon.
	NeonAPIKey string `env:"NEON_API_KEY,required"`

	// DebugDB enables debug mode for the database.
	DebugDB bool `env:"DB_DEBUG" envDefault:"false"`

	// All queries will use these filters. One example of the default filter is:
	// regions.provider = $PROVIDER
	// That filter is always enabled.
	RegionFilters string `env:"REGION_FILTERS"`
}

func ParseEnv() (*App, error) {
	cfg := App{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
