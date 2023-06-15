package rules

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
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
	queryRepo   *repos.QueryRepo
	neonClient  *neonapi.Client
	register    *bgjobs.Register
	exitnode    string
}

type DeleteProjectArgs struct {
	N int
}

func NewDeleteProject(a *app.App, j json.RawMessage) (*DeleteProject, error) {
	var args DeleteProjectArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	return &DeleteProject{
		projectsN:   args.N,
		provider:    a.Config.Provider,
		regionRepo:  a.Repo.Region,
		projectRepo: a.Repo.Project,
		queryRepo:   a.Repo.Query,
		neonClient:  a.NeonClient,
		register:    a.Register,
		exitnode:    a.Config.Exitnode,
	}, nil
}

func (c *DeleteProject) Execute(ctx context.Context) error {
	regions, err := c.regionRepo.FindByProvider(c.provider)
	if err != nil {
		return err
	}

	for _, region := range regions {
		c.register.Go(func() { c.executeForRegion(ctx, region) })
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
	// TODO: kill background jobs for this project and wait for them to finish

	// preparing a query
	prep, err := c.neonClient.DeleteProject(projectDB.ProjectID)
	if err != nil {
		return err
	}

	dbQuery := prep.Query(&projectDB.ID, projectDB.RegionID, c.exitnode)

	// 1. save delete_project query to db
	err = c.queryRepo.Save(dbQuery)
	if err != nil {
		return fmt.Errorf("failed to perist delete_project query: %w", err)
	}

	// 2. set deleted_at in db
	err = c.projectRepo.Delete(projectDB)
	if err != nil {
		return err
	}

	// 3. call API
	_, result, err := prep.Do(ctx)

	dbErr := c.queryRepo.FinishSaveResult(dbQuery, result)
	if dbErr != nil {
		log.Error(ctx, "failed to persist query result", zap.Error(dbErr))
		if err == nil {
			err = dbErr
		} else {
			err = errors.Join(err, dbErr)
		}
	}
	if err != nil {
		return err
	}

	log.Info(ctx, "project deleted")
	return nil
}
