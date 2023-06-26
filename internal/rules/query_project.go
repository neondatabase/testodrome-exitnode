package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/drivers"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

// Rule to send random queries to random projects.
type QueryProject struct {
	args          QueryProjectArgs
	regionFilters []repos.Filter
	regionRepo    *repos.RegionRepo
	projectRepo   *repos.ProjectRepo
	queryRepo     *repos.QueryRepo
	register      *bgjobs.Register
	exitnode      string
	projectLocker *bgjobs.ProjectLocker
}

type QueryProjectArgs struct {
	UsePooler rdesc.Wrand[bool]
}

var defaultUsePooler = rdesc.Wrand[bool]{
	{Weight: 1, Item: true},
	{Weight: 1, Item: false},
}

func NewQueryProject(a *app.App, j json.RawMessage) (*QueryProject, error) {
	var args QueryProjectArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	if args.UsePooler == nil {
		args.UsePooler = defaultUsePooler
	}

	return &QueryProject{
		args:          args,
		regionFilters: a.RegionFilters,
		regionRepo:    a.Repo.Region,
		projectRepo:   a.Repo.Project,
		queryRepo:     a.Repo.Query,
		register:      a.Register,
		exitnode:      a.Config.Exitnode,
		projectLocker: a.ProjectLocker,
	}, nil
}

func (r *QueryProject) Execute(ctx context.Context) error {
	regions, err := r.regionRepo.Find(r.regionFilters)
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

	filters := []repos.Filter{
		repos.FilterByRegionID(region.ID),
	}
	projects, err := r.projectRepo.FindRandomProjects(filters, 1)
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

func appendPoolerSuffix(connstr string) (string, error) {
	// connstr has `@<endpoint_id>.` substring in it
	// we need to replace it with `@<endpoint_id>-pooler.`
	// using regex for that
	re := regexp.MustCompile(`@([a-z0-9\-]+)\.`)
	newConnstr := re.ReplaceAllString(connstr, "@$1-pooler.")
	if newConnstr == connstr {
		return "", fmt.Errorf("failed to replace connstr with pooler")
	}
	return newConnstr, nil
}

func (r *QueryProject) randomDriver(ctx context.Context, project models.Project, usePooler bool) (drivers.Driver, error) {
	// save queries to the database with project and exitnode info
	saver := repos.NewQuerySaver(r.queryRepo, repos.QuerySaverArgs{
		ProjectID: &project.ID,
		Exitnode:  &r.exitnode,
		RegionID:  &project.RegionID,
	})

	connstr := project.ConnectionString
	if usePooler {
		var err error
		connstr, err = appendPoolerSuffix(connstr)
		if err != nil {
			return nil, fmt.Errorf("failed to append pooler suffix: %w", err)
		}
	}

	num := rand.Intn(3)
	switch num {
	case 0:
		log.Info(ctx, "using serverless driver")
		return drivers.NewServerless(connstr, saver)
	case 1:
		log.Info(ctx, "using vercel-sl driver")
		return drivers.NewVercelSL(connstr, saver), nil
	case 2:
		log.Info(ctx, "using pgx driver")
		return drivers.PgxConnect(ctx, connstr, saver)
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

	usePooler := r.args.UsePooler.Pick()

	// TODO: use a library to select random query and random driver

	driver, err1 := r.randomDriver(ctx, project, usePooler)
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
