package rules

import (
	"fmt"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/conf"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	log "github.com/sirupsen/logrus"
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
}

func NewCreateProject(a *app.App, interval time.Duration) *CreateProject {
	return &CreateProject{
		interval:    interval,
		provider:    a.Config.Provider,
		regionRepo:  a.Repo.Region,
		projectRepo: a.Repo.Project,
		sequence:    a.Repo.SeqExitnodeProject,
		neonClient:  a.NeonClient,
		config:      a.Config,
	}
}

func (c *CreateProject) Execute() error {
	regions, err := c.regionRepo.FindByProvider(c.provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
		go c.executeForRegion(region)
	}
	return nil
}

// Execute rule for a single region. Will create a project only if the last created project
// is older than GapDuration.
func (c *CreateProject) executeForRegion(region models.Region) {
	logger := log.WithField("regionID", region.ID)

	project, err := c.projectRepo.FindLastCreatedProject(region.ID)
	if err != nil {
		logger.WithError(err).Error("failed to find last created project")
		return
	}

	if project == nil || time.Since(project.CreatedAt) > c.interval {
		logger.Info("creating project")
		err := c.createProject(region)
		if err != nil {
			logger.WithError(err).Error("failed to create project")
			return
		}
	}
}

// Create a project in the given region.
func (c *CreateProject) createProject(region models.Region) error {
	projectSeqID, err := c.sequence.Next()
	if err != nil {
		return err
	}

	// TODO: store information about project creation API query in the database.
	createRequest := &neonapi.CreateProject{
		Name:     fmt.Sprintf("test@%s-%d", c.config.Exitnode, projectSeqID),
		RegionID: region.DatabaseRegion,
	}

	project, err := c.neonClient.CreateProject(createRequest)
	if err != nil {
		return err
	}

	var connstr string
	if len(project.ConnectionUris) == 1 {
		connstr = project.ConnectionUris[0].ConnectionURI
	} else {
		log.WithField("projectID", project.Project.ID).Warn("project has invalid number of connection strings")
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
