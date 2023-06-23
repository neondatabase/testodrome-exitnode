package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/drivers"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

// Rule to send random queries to random projects.
type QueryProject struct {
	// Projects will be queried only in regions with this provider.
	// TODO: make custom filters already
	provider      string
	regionRepo    *repos.RegionRepo
	projectRepo   *repos.ProjectRepo
	queryRepo     *repos.QueryRepo
	register      *bgjobs.Register
	exitnode      string
	projectLocker *bgjobs.ProjectLocker
}

type QueryProjectArgs struct {
}

func NewQueryProject(a *app.App, j json.RawMessage) (*QueryProject, error) {
	var args QueryProjectArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	return &QueryProject{
		provider:      a.Config.Provider,
		regionRepo:    a.Repo.Region,
		projectRepo:   a.Repo.Project,
		queryRepo:     a.Repo.Query,
		register:      a.Register,
		exitnode:      a.Config.Exitnode,
		projectLocker: a.ProjectLocker,
	}, nil
}

func (r *QueryProject) Execute(ctx context.Context) error {
	regions, err := r.regionRepo.FindByProvider(r.provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
		region := region
		r.register.Go(func() { r.executeForRegion(ctx, region) })
	}
	return nil
}

// Execute rule for a single region. Will delete a project only if the are too many.
func (r *QueryProject) executeForRegion(ctx context.Context, region models.Region) {
	ctx = log.With(ctx, zap.Uint("regionID", region.ID))

	projects, err := r.projectRepo.FindRandomProjects(region.ID, 1)
	if err != nil {
		log.Error(ctx, "failed to find random project", zap.Error(err))
		return
	}

	for _, p := range projects {
		err := r.executeForProject(ctx, p)
		if err != nil {
			log.Error(ctx, "failed to execute queries for project", zap.Error(err), zap.Uint("projectID", p.ID))
		}
	}
}

func (r *QueryProject) randomDriver(ctx context.Context, project models.Project) (drivers.Driver, error) {
	// save queries to the database with project and exitnode info
	saver := repos.NewQuerySaver(r.queryRepo, repos.QuerySaverArgs{
		ProjectID: &project.ID,
		Exitnode:  &r.exitnode,
		RegionID:  &project.RegionID,
	})

	num := rand.Intn(3)
	switch num {
	case 0:
		log.Info(ctx, "using serverless driver")
		return drivers.NewServerless(project.ConnectionString, saver)
	case 1:
		log.Info(ctx, "using vercel-sl driver")
		return drivers.NewVercelSL(project.ConnectionString, saver), nil
	case 2:
		log.Info(ctx, "using pgx driver")
		return drivers.PgxConnect(ctx, project.ConnectionString, saver)
	}

	panic("unreachable")
}

// Execute a random query for a single project.
func (r *QueryProject) executeForProject(ctx context.Context, project models.Project) error {
	ctx = log.With(ctx, zap.Uint("projectID", project.ID))
	projectLock := r.projectLocker.Get(project.ID)

	unlock := projectLock.TrySharedLock()
	if unlock == nil {
		return fmt.Errorf("project is locked")
	}
	defer unlock()

	// TODO: use a library to select random query and random driver

	driver, err1 := r.randomDriver(ctx, project)
	if err1 != nil {
		return fmt.Errorf("failed to create driver: %w", err1)
	}

	const select1 = `SELECT 1`
	const createTable = `CREATE TABLE IF NOT EXISTS activity_v1 (
		id SERIAL PRIMARY KEY,
		nonce BIGINT,
		val FLOAT,
		created_at TIMESTAMP DEFAULT NOW()
	  )`
	const doActivity = `INSERT INTO activity_v1(nonce,val) SELECT $1 AS nonce, avg(id) AS val FROM activity_v1 RETURNING *`

	queries := []drivers.SingleQuery{
		// first query, can trigger a cold start
		{Query: select1},
		// init table
		{Query: createTable},
		// do some activity
		{Query: doActivity, Params: []any{rand.Int63()}},
	}

	var res []models.Query
	var err error
	if md, ok := driver.(drivers.ManyQueriesDriver); ok {
		res, err = md.Queries(ctx, queries...)
	} else {
		for _, q := range queries {
			query, err2 := driver.Query(ctx, q)
			if query != nil {
				res = append(res, *query)
			}
			if err2 != nil {
				err = err2
				break
			}
		}
	}

	_ = res
	return err
}
