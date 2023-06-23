package rules

import (
	"context"
	"encoding/json"
	"errors"
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
	provider    string
	regionRepo  *repos.RegionRepo
	projectRepo *repos.ProjectRepo
	queryRepo   *repos.QueryRepo
	register    *bgjobs.Register
	exitnode    string
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
		provider:    a.Config.Provider,
		regionRepo:  a.Repo.Region,
		projectRepo: a.Repo.Project,
		queryRepo:   a.Repo.Query,
		register:    a.Register,
		exitnode:    a.Config.Exitnode,
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

// Execute a random query for a single project.
func (r *QueryProject) executeForProject(ctx context.Context, project models.Project) error {
	ctx = log.With(ctx, zap.Uint("projectID", project.ID))
	// TODO: use a library to select random query and random driver

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

	driver := rand.Intn(2)
	if driver == 0 {
		return r.goServerlessDriver(ctx, project, queries)
	} else {
		return r.vercelSLDriver(ctx, project, queries)
	}
}

func (r *QueryProject) goServerlessDriver(ctx context.Context, project models.Project, queries []drivers.SingleQuery) error {
	driver, err1 := drivers.NewServerless(r.exitnode, &project)
	if err1 != nil {
		return err1
	}

	for _, q := range queries {
		query, err2 := driver.Query(ctx, q.Query, q.Params...)
		if err := r.saveQuery(query, err2); err != nil {
			return err
		}
	}

	return nil
}

func (r *QueryProject) vercelSLDriver(ctx context.Context, project models.Project, queries []drivers.SingleQuery) error {
	driver := drivers.NewVercelSL(&project)
	res, err2 := driver.Queries(ctx, queries...)

	for _, q := range res {
		q := q
		if err := r.saveQuery(&q, err2); err != nil {
			return err
		}
	}

	return err2
}

func (r *QueryProject) saveQuery(query *models.Query, queryErr error) (retErr error) {
	retErr = queryErr

	if err := r.queryRepo.Save(query); err != nil {
		if retErr == nil {
			retErr = err
		} else {
			retErr = errors.Join(retErr, err)
		}
	}

	return retErr
}
