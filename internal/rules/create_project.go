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
	// Projects will be created only in regions with this provider.
	provider    string
	regionRepo  *repos.RegionRepo
	projectRepo *repos.ProjectRepo
	sequence    *repos.Sequence
	neonClient  *neonapi.Client
	config      *conf.App
	register    *bgjobs.Register
}

type CreateProjectArgs struct {
	Interval rdesc.Duration
}

func NewCreateProject(a *app.App, j json.RawMessage) (*CreateProject, error) {
	var args CreateProjectArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	return &CreateProject{
		interval:    args.Interval.Duration,
		provider:    a.Config.Provider,
		regionRepo:  a.Repo.Region,
		projectRepo: a.Repo.Project,
		sequence:    a.Repo.SeqExitnodeProject,
		neonClient:  a.NeonClient,
		config:      a.Config,
		register:    a.Register,
	}, nil
}

func (c *CreateProject) Execute(ctx context.Context) error {
	regions, err := c.regionRepo.FindByProvider(c.provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
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

	// TODO: store information about project creation API query in the database.
	createRequest := &neonapi.CreateProject{
		Name:     fmt.Sprintf("test@%s-%d", c.config.Exitnode, projectSeqID),
		RegionID: region.DatabaseRegion,
	}

	project, err := c.neonClient.CreateProject(ctx, createRequest)
	if err != nil {
		return err
	}
	ctx = log.With(ctx, zap.String("projectID", project.Project.ID))

	var connstr string
	if len(project.ConnectionUris) == 1 {
		connstr = project.ConnectionUris[0].ConnectionURI
	} else {
		log.Warn(ctx, "project has invalid number of connection strings")
	}

	dbProject := models.Project{
		RegionID:          region.ID,
		Name:              project.Project.Name,
		ProjectID:         project.Project.ID,
		ConnectionString:  connstr,
		CreatedByExitnode: c.config.Exitnode,
	}

	err = c.projectRepo.Create(&dbProject)
	if err != nil {
		return fmt.Errorf("failed to create project in the database: %w", err)
	}

	return nil
}
