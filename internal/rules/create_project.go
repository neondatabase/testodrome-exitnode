package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/conf"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

// Rule to create a project in every region at least once per `interval` minutes.
type CreateProject struct {
	interval time.Duration
	args     CreateProjectArgs
	// Projects will be created only in regions adhering to these filters.
	regionFilters []repos.Filter
	regionRepo    *repos.RegionRepo
	projectRepo   *repos.ProjectRepo
	queryRepo     *repos.QueryRepo
	sequence      *repos.Sequence
	neonClient    *neonapi.Client
	config        *conf.App
	register      *bgjobs.Register
}

type CreateProjectArgs struct {
	Interval       rdesc.Duration
	PgVersion      rdesc.Wrand[int]
	Provisioner    rdesc.Wrand[string]
	SuspendTimeout rdesc.Wrand[int]
	Mode           rdesc.Wrand[string]
}

var defaultPgVersion = rdesc.Wrand[int]{
	{Weight: 1, Item: 15},
	{Weight: 1, Item: 14},
}

var defaultProvisioner = rdesc.Wrand[string]{
	{Weight: 1, Item: "k8s-pod"},
	{Weight: 1, Item: "k8s-neonvm"},
}

const stdProvisioner = "k8s-pod"

var defaultSuspendTimeout = rdesc.Wrand[int]{
	// 0 means default timeout.
	{Weight: 20, Item: 0},
	// 1 second timeout.
	{Weight: 1, Item: 1},
}

var defaultProjectMode = rdesc.Wrand[string]{
	{Weight: 1, Item: ""},
}

func NewCreateProject(a *app.App, j json.RawMessage) (*CreateProject, error) {
	var args CreateProjectArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	if args.PgVersion == nil {
		args.PgVersion = defaultPgVersion
	}
	if args.Provisioner == nil {
		args.Provisioner = defaultProvisioner
	}
	if args.SuspendTimeout == nil {
		args.SuspendTimeout = defaultSuspendTimeout
	}
	if args.Mode == nil {
		args.Mode = defaultProjectMode
	}

	return &CreateProject{
		interval:      args.Interval.Duration,
		args:          args,
		regionFilters: a.RegionFilters,
		regionRepo:    a.Repo.Region,
		projectRepo:   a.Repo.Project,
		queryRepo:     a.Repo.Query,
		sequence:      a.Repo.SeqExitnodeProject,
		neonClient:    a.NeonClient,
		config:        a.Config,
		register:      a.Register,
	}, nil
}

func (c *CreateProject) Execute(ctx context.Context) error {
	regions, err := c.regionRepo.Find(c.regionFilters)
	if err != nil {
		return err
	}

	for _, region := range regions {
		region := region
		c.register.Go(func() { c.executeForRegion(ctx, region) })
	}
	return nil
}

// Execute rule for a single region. Will create a project only if the last created project
// is older than GapDuration.
func (c *CreateProject) executeForRegion(ctx context.Context, region models.Region) {
	ctx = log.With(ctx, zap.Uint("regionID", region.ID))

	project, err := c.projectRepo.FindLastCreatedProject(region.ID)
	if err != nil {
		log.Error(ctx, "failed to find last created project", zap.Error(err))
		return
	}

	if project == nil || time.Since(project.CreatedAt) > c.interval {
		log.Info(ctx, "creating project")
		err := c.createProject(ctx, region)
		if err != nil {
			log.Error(ctx, "failed to create project", zap.Error(err))
			return
		}
	}
}

// Create a project in the given region.
func (c *CreateProject) createProject(ctx context.Context, region models.Region) error {
	projectSeqID, err := c.sequence.Next()
	if err != nil {
		return err
	}

	provisioner := c.args.Provisioner.Pick()
	if !region.SupportsNeonVM {
		provisioner = stdProvisioner
	}

	suspendTimeout := c.args.SuspendTimeout.Pick()

	createRequest := &neonapi.CreateProject{
		Name:        fmt.Sprintf("test@%s-%d", c.config.Exitnode, projectSeqID),
		RegionID:    region.DatabaseRegion,
		PgVersion:   c.args.PgVersion.Pick(),
		Provisioner: provisioner,
	}

	prep, err := c.neonClient.CreateProject(createRequest)
	if err != nil {
		return err
	}

	saver := repos.NewQuerySaver(c.queryRepo, repos.QuerySaverArgs{
		ProjectID: nil,
		RegionID:  &region.ID,
		Exitnode:  &c.config.Exitnode,
	})

	project, err := queryAPI(ctx, prep, saver)
	if err != nil {
		return err
	}

	ctx = log.With(ctx, zap.String("projectID", project.Project.ID))

	if err2 := c.postCreate(ctx, saver, project, suspendTimeout); err2 != nil {
		log.Error(ctx, "failed post create", zap.Error(err2))
		return err2
	}

	var connstr string
	if len(project.ConnectionUris) == 1 {
		connstr = project.ConnectionUris[0].ConnectionURI
	} else {
		log.Warn(ctx, "project has invalid number of connection strings")
	}

	mode := c.args.Mode.Pick()

	dbProject := models.Project{
		RegionID:              region.ID,
		Name:                  project.Project.Name,
		ProjectID:             project.Project.ID,
		ConnectionString:      connstr,
		CreatedByExitnode:     c.config.Exitnode,
		PgVersion:             project.Project.PgVersion,
		Provisioner:           project.Project.Provisioner,
		SuspendTimeoutSeconds: suspendTimeout,
		CurrentMode:           mode,
	}

	err = c.projectRepo.Create(&dbProject)
	if err != nil {
		return fmt.Errorf("failed to create project in the database: %w", err)
	}

	return nil
}

func (c *CreateProject) postCreate(
	ctx context.Context,
	saver *repos.QuerySaver,
	project *neonapi.CreateProjectResponse,
	suspendTimeout int,
) error {
	if len(project.Endpoints) != 1 {
		log.Warn(ctx, "project has invalid number of endpoints", zap.Any("endpoints", project.Endpoints))
		return nil
	}
	endpoint := project.Endpoints[0]

	if suspendTimeout == endpoint.SuspendTimeoutSeconds {
		return nil
	}

	// wait until all operations are finished, otherwise we will get an error:
	// `project already has running operations, scheduling of new ones is prohibited`
	if err := c.waitAllOperations(ctx, saver, project); err != nil {
		return err
	}

	log.Info(ctx, "updating suspend timeout", zap.Int("old", endpoint.SuspendTimeoutSeconds), zap.Int("new", suspendTimeout))
	prep, err := c.neonClient.UpdateEndpoint(project.Project.ID, endpoint.ID, &neonapi.UpdateEndpoint{
		SuspendTimeoutSeconds: &suspendTimeout,
	})
	if err != nil {
		return err
	}

	_, err = queryAPI(ctx, prep, saver)
	return err
}

func (c *CreateProject) waitAllOperations(ctx context.Context, saver *repos.QuerySaver, project *neonapi.CreateProjectResponse) error {
	// project.Operations has a list of running operations, but it's not convenient to poll them one by one
	prep, err := c.neonClient.GetOperations(project.Project.ID)
	if err != nil {
		return err
	}

	const startInterval = 2 * time.Second
	const maxInterval = 60 * time.Second

	sleepInterval := startInterval

	for {
		ops, err := queryAPI(ctx, prep, saver)
		if err != nil {
			return err
		}

		allFinished := true

		for _, op := range ops.Operations {
			if op.Status != "finished" {
				log.Debug(ctx, "operation not finished", zap.Any("operation", op))
				allFinished = false
				break
			}
		}

		if allFinished {
			break
		}

		time.Sleep(sleepInterval)
		sleepInterval += sleepInterval / 2
		if sleepInterval > maxInterval {
			sleepInterval = maxInterval
		}
	}

	return nil
}
