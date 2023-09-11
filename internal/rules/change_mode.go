package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/bgjobs"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
	"github.com/petuhovskiy/neon-lights/internal/rdesc"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

type ChangeMode struct {
	args           ChangeModeArgs
	projectFilters []repos.Filter
	projectRepo    *repos.ProjectRepo
	register       *bgjobs.Register
	projectLocker  *bgjobs.ProjectLocker
	queryProject   *QueryProject
}

type ChangeModeArgs struct {
	NewMode          rdesc.Wrand[string]
	RawProjectFilter string
	// This is QueryProjectArgs
	QueryBeforeChange json.RawMessage
}

func NewChangeMode(a *app.App, j json.RawMessage) (*ChangeMode, error) {
	var args ChangeModeArgs
	err := json.Unmarshal(j, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	if args.NewMode == nil {
		return nil, fmt.Errorf("NewMode field must be set")
	}

	var projectFilters []repos.Filter
	projectFilters = append(projectFilters, a.RegionFilters...)
	if args.RawProjectFilter != "" {
		projectFilters = append(projectFilters, repos.RawFilter(args.RawProjectFilter))
	}

	queryProject, err := NewQueryProject(a, args.QueryBeforeChange)
	if err != nil {
		return nil, fmt.Errorf("failed to NewQueryProject: %w", err)
	}

	return &ChangeMode{
		args:           args,
		projectFilters: projectFilters,
		projectRepo:    a.Repo.Project,
		register:       a.Register,
		projectLocker:  a.ProjectLocker,
		queryProject:   queryProject,
	}, nil
}

func (r *ChangeMode) Execute(ctx context.Context) error {
	projects, err := r.projectRepo.FindRandomProjects(r.projectFilters, 1)
	if err != nil {
		return fmt.Errorf("failed to find random project: %w", err)
	}

	for _, project := range projects {
		r.startExecuteProject(ctx, project)
	}

	return nil
}

func (r *ChangeMode) startExecuteProject(ctx context.Context, project models.Project) {
	newMode := r.args.NewMode.Pick()

	ctx = log.With(ctx, zap.Uint("projectID", project.ID))
	ctx = log.With(ctx, zap.String("newMode", newMode))

	r.register.Go(func() {
		err := r.executeForProject(ctx, project, newMode)
		if err != nil {
			log.Error(ctx, "failed to execute project", zap.Error(err))
		}
	})
}

// Execute a random query for a single project.
func (r *ChangeMode) executeForProject(ctx context.Context, project models.Project, newMode string) error {
	projectLock := r.projectLocker.Get(project.ID)

	// start with querying a project
	err := r.queryProject.executeForProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to query project before changing mode: %w", err)
	}

	// try to take a lock
	unlock := projectLock.TryExclusiveLock()
	if unlock == nil {
		return ErrProjectLocked
	}
	defer unlock()

	// project lock is taken now
	log.Info(ctx, "updating project mode", zap.String("prevMode", project.CurrentMode))

	err = r.projectRepo.UpdateMode(&project, newMode)
	if err != nil {
		return err
	}

	// TODO: race condition with query project is possible
	time.Sleep(time.Second)

	return nil
}
