package rules

import (
	"fmt"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	"github.com/petuhovskiy/neon-lights/pkg/conf"
	log "github.com/sirupsen/logrus"
)

// TODO: refactor to don't use public fields.

// Rule to create a project in every region at least once per GapDuration minutes.
type CreateProject struct {
	GapDuration time.Duration
	// Projects will be created only in regions with this provider.
	Provider    string
	RegionRepo  *repos.RegionRepo
	ProjectRepo *repos.ProjectRepo
	Sequence    *repos.Sequence
	NeonClient  *neonapi.Client
	Config      *conf.App
}

func (c *CreateProject) Execute() error {
	regions, err := c.RegionRepo.FindByProvider(c.Provider)
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

	project, err := c.ProjectRepo.FindLastCreatedProject(region.ID)
	if err != nil {
		logger.WithError(err).Error("failed to find last created project")
		return
	}

	if project == nil || time.Since(project.CreatedAt) > c.GapDuration {
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
	projectSeqID, err := c.Sequence.Next()
	if err != nil {
		return err
	}

	// TODO: store information about project creation API query in the database.
	createRequest := &neonapi.CreateProject{
		Name:     fmt.Sprintf("@%s-%d", c.Config.Exitnode, projectSeqID),
		RegionID: region.DatabaseRegion,
	}

	project, err := c.NeonClient.CreateProject(createRequest)
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
		CreatedByExitnode: c.Config.Exitnode,
	}

	err = c.ProjectRepo.Create(&dbProject)
	if err != nil {
		return fmt.Errorf("failed to create project in the database: %w", err)
	}

	return nil
}
