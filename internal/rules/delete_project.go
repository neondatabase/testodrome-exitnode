package rules

import (
	"math/rand"
	"sort"

	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	log "github.com/sirupsen/logrus"
)

// TODO: refactor to don't use public fields.

// Rule to delete random projects when there are too many projects with the similar configuration (matrix).
// TODO: make it work with custom matrix, not only per-region.
type DeleteProject struct {
	// Target number of projects. Project will be deleted if there are more than this number of projects.
	ProjectsN int
	// Projects will be deleted only in regions with this provider.
	Provider    string
	RegionRepo  *repos.RegionRepo
	ProjectRepo *repos.ProjectRepo
	NeonClient  *neonapi.Client
}

func (c *DeleteProject) Execute() error {
	regions, err := c.RegionRepo.FindByProvider(c.Provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
		go c.executeForRegion(region)
	}
	return nil
}

// Execute rule for a single region. Will delete a project only if the are too many.
func (c *DeleteProject) executeForRegion(region models.Region) {
	logger := log.WithField("regionID", region.ID)

	// TODO: it is not efficient to load all projects here, but it is ok for now.
	projects, err := c.ProjectRepo.FindAllByRegion(region.ID)
	if err != nil {
		logger.WithError(err).Error("failed to find projects")
		return
	}

	if len(projects) <= c.ProjectsN {
		return
	}

	// shuffle projects to delete random ones
	rand.Shuffle(len(projects), func(i, j int) {
		projects[i], projects[j] = projects[j], projects[i]
	})

	// take any N projects
	projects = projects[:c.ProjectsN]

	// sort by creation date
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.Before(projects[j].CreatedAt)
	})

	// take the middle project, because we don't want to take too old and too new projects
	project := projects[len(projects)/2]
	log.WithField("projectID", project.ID).Info("selected project for deletion")

	err = c.deleteProject(&project)
	if err != nil {
		logger.WithError(err).Error("failed to delete project")
		return
	}
}

// Delete a project.
func (c *DeleteProject) deleteProject(projectDB *models.Project) error {
	// calling a delete API
	err := c.NeonClient.DeleteProject(projectDB.ProjectID)
	if err != nil {
		// TODO: retry, otherwise state will be inconsistent
		return err
	}

	err = c.ProjectRepo.Delete(projectDB)
	if err != nil {
		return err
	}

	log.WithField("projectID", projectDB.ID).Info("project deleted")
	return nil
}
