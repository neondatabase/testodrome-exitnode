package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/drivers"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

var ErrConcurrencyLimit = fmt.Errorf("concurrency limit reached")
var ErrProjectLocked = fmt.Errorf("project locked")

// Rule to send random queries to random projects.
type QueryProject struct {
	args           QueryProjectArgs
	projectFilters []repos.Filter
	regionRepo     *repos.RegionRepo
	projectRepo    *repos.ProjectRepo
	queryRepo      *repos.QueryRepo
	register       *bgjobs.Register
	exitnode       string
	projectLocker  *bgjobs.ProjectLocker
	scenario       queryScenario
	nowRunning     atomic.Int64
}

type QueryProjectArgs struct {
	ConcurrencyLimit  int
	Scenario          string
	UsePooler         rdesc.Wrand[bool]
	Driver            rdesc.Wrand[drivers.Name]
	MaxRandomProjects uint
	RawProjectFilter  string
}

var defaultUsePooler = rdesc.Wrand[bool]{
	{Weight: 1, Item: true},
	{Weight: 1, Item: false},
}

var defaultDrivers = rdesc.Wrand[drivers.Name]{
	{Weight: 1, Item: drivers.PgxConn},
	{Weight: 1, Item: drivers.GoServerless},
	{Weight: 1, Item: drivers.VercelEdge},
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

	if args.Driver == nil {
		args.Driver = defaultDrivers
	}

	if args.MaxRandomProjects < 1 {
		args.MaxRandomProjects = 1
	}

	scenario, err := getScenario(args.Scenario)
	if err != nil {
		return nil, err
	}

	var projectFilters []repos.Filter
	projectFilters = append(projectFilters, a.RegionFilters...)
	if args.RawProjectFilter != "" {
		projectFilters = append(projectFilters, repos.RawFilter(args.RawProjectFilter))
	}

	return &QueryProject{
		args:           args,
		projectFilters: projectFilters,
		regionRepo:     a.Repo.Region,
		projectRepo:    a.Repo.Project,
		queryRepo:      a.Repo.Query,
		register:       a.Register,
		exitnode:       a.Config.Exitnode,
		projectLocker:  a.ProjectLocker,
		scenario:       scenario,
	}, nil
}

func (r *QueryProject) Execute(ctx context.Context) error {
	projects, err := r.projectRepo.FindRandomProjects(r.projectFilters, int(r.args.MaxRandomProjects))
	if err != nil {
		return fmt.Errorf("failed to find random project: %w", err)
	}

	for _, project := range projects {
		r.startExecuteProject(ctx, project)
	}

	return nil
}

func (r *QueryProject) startExecuteProject(ctx context.Context, project models.Project) {
	ctx = log.With(ctx, zap.Uint("projectID", project.ID))
	ctx = log.With(ctx, zap.String("scenario", r.args.Scenario))

	r.register.Go(func() {
		err := r.executeForProject(ctx, project)
		if err == ErrConcurrencyLimit || err == ErrProjectLocked {
			err = nil
		}
		if err != nil {
			log.Error(ctx, "failed to execute project", zap.Error(err))
		}
	})
}

// Execute a random query for a single project.
func (r *QueryProject) executeForProject(ctx context.Context, project models.Project) error {
	projectLock := r.projectLocker.Get(project.ID)

	usePooler := r.args.UsePooler.Pick()

	scenario := r.scenario
	needExclusiveLock := scenario.exclusive()

	var unlock func()
	if needExclusiveLock {
		unlock = projectLock.TryExclusiveLock()
	} else {
		unlock = projectLock.TrySharedLock()
	}

	if unlock == nil {
		return ErrProjectLocked
	}
	defer unlock()

	// project lock is taken now

	// checking concurrency level
	atmc := r.nowRunning.Add(1)
	defer r.nowRunning.Add(-1)

	if r.args.ConcurrencyLimit > 0 && atmc > int64(r.args.ConcurrencyLimit) {
		return ErrConcurrencyLimit
	}

	// running the scenario
	driver, err1 := r.randomDriver(ctx, project, usePooler)
	if err1 != nil {
		return fmt.Errorf("failed to create driver: %w", err1)
	}

	if cc, ok := driver.(drivers.CloseableDriver); ok {
		defer cc.Close(ctx)
	}

	return scenario.execute(ctx, queryParams{
		driver:  driver,
		project: &project,
	})
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

	driverName := r.args.Driver.Pick()
	connstr += fmt.Sprintf("?application_name=testodrome/%s", string(driverName))

	switch driverName {
	case drivers.PgxConn:
		log.Info(ctx, "using pgx driver")
		return drivers.PgxConnect(ctx, connstr, saver)
	case drivers.GoServerless:
		log.Info(ctx, "using serverless driver")
		return drivers.NewServerless(connstr, saver)
	case drivers.VercelEdge:
		log.Info(ctx, "using vercel-sl driver")
		return drivers.NewVercelSL(connstr, saver), nil
	}

	return nil, fmt.Errorf("unknown driver: %s", driverName)
}
