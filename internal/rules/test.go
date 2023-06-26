package rules

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/app"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

type TestRule struct {
	regionFilters []repos.Filter
	projectRepo   *repos.ProjectRepo
	matrix        []string
}

func NewTestRule(a *app.App, _ json.RawMessage) (*TestRule, error) {
	return &TestRule{
		projectRepo:   a.Repo.Project,
		regionFilters: a.RegionFilters,
		matrix:        defaultMatrix,
	}, nil
}

func (r *TestRule) Execute(ctx context.Context) error {
	p1, err := r.projectRepo.FindRandomProjects(r.regionFilters, 1)
	if err != nil {
		return err
	}

	if len(p1) != 1 {
		log.Info(ctx, "no projects found")
		return nil
	}

	project := p1[0]
	log.Info(ctx, "test rule", zap.Any("project", project))

	filters, err := repos.MatrixFilters(&project, r.matrix)
	if err != nil {
		return err
	}
	filters = append(filters, r.regionFilters...)

	const randomNumber = 5
	projects, err := r.projectRepo.FindRandomProjects(filters, randomNumber)
	if err != nil {
		return err
	}

	log.Info(ctx, "selected random", zap.Any("projects", projects), zap.Any("matrix", r.matrix))
	return nil
}
