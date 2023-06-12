package rules

import (
	"context"
	"math/rand"
	"sort"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
	"go.uber.org/zap"
)

// Rule to delete random projects when there are too many projects with the similar configuration (matrix).
// TODO: make it work with custom matrix, not only per-region.
type DeleteProject struct {
	// Target number of projects. Project will be deleted if there are more than this number of projects.
	projectsN int
	// Projects will be deleted only in regions with this provider.
	provider    string
	regionRepo  *repos.RegionRepo
	projectRepo *repos.ProjectRepo
	neonClient  *neonapi.Client
}

func NewDeleteProject(a *app.App, projectsN int) *DeleteProject {
	return &DeleteProject{
		projectsN:   projectsN,
		provider:    a.Config.Provider,
		regionRepo:  a.Repo.Region,
		projectRepo: a.Repo.Project,
		neonClient:  a.NeonClient,
	}
}

func (c *DeleteProject) Execute(ctx context.Context) error {
	ctx = log.With(ctx, zap.String("rule", "delete_project"))

	regions, err := c.regionRepo.FindByProvider(c.provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
		go c.executeForRegion(ctx, region)
	}
	return nil
}

// Execute rule for a single region. Will delete a project only if the are too many.
func (c *DeleteProject) executeForRegion(ctx context.Context, region models.Region) {
	ctx = log.With(ctx, zap.Uint("regionID", region.ID))

	// TODO: it is not efficient to load all projects here, but it is ok for now.
	projects, err := c.projectRepo.FindAllByRegion(region.ID)
	if err != nil {
		log.Error(ctx, "failed to load projects", zap.Error(err))
		return
	}

	if len(projects) <= c.projectsN {
		return
	}

	// shuffle projects to delete random ones
	rand.Shuffle(len(projects), func(i, j int) {
		projects[i], projects[j] = projects[j], projects[i]
	})

	// take any N projects
	projects = projects[:c.projectsN]

	// sort by creation date
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.Before(projects[j].CreatedAt)
	})

	// take the middle project, because we don't want to take too old and too new projects
	project := projects[len(projects)/2]
	ctx = log.With(ctx, zap.Uint("projectID", project.ID))
	log.Info(ctx, "selected project for deletion")

	err = c.deleteProject(ctx, &project)
	if err != nil {
		log.Error(ctx, "failed to delete project", zap.Error(err))
		return
	}
}

// Delete a project.
func (c *DeleteProject) deleteProject(ctx context.Context, projectDB *models.Project) error {
	// calling a delete API
	err := c.neonClient.DeleteProject(ctx, projectDB.ProjectID)
	if err != nil {
		// TODO: retry, otherwise state will be inconsistent
		return err
	}

	err = c.projectRepo.Delete(projectDB)
	if err != nil {
		return err
	}

	log.Info(ctx, "project deleted")
	return nil
}
