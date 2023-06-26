package repos

import (
	"gorm.io/gorm"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{
		db: db,
	}
}

// FindLastCreatedProject returns the last created project in the region.
// May return deleted projects.
func (r *ProjectRepo) FindLastCreatedProject(regionID uint) (*models.Project, error) {
	var projects []models.Project
	err := r.db.
		Unscoped().
		Where("region_id = ?", regionID).
		Order("created_at DESC").
		Limit(1).
		Find(&projects).
		Error
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, nil
	}
	return &projects[0], nil
}

func (r *ProjectRepo) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepo) FindAllByRegion(regionID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.
		Where("region_id = ?", regionID).
		Find(&projects).
		Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepo) Delete(project *models.Project) error {
	return r.db.Delete(project).Error
}

func (r *ProjectRepo) FindRandomProjects(filters []Filter, n int) ([]models.Project, error) {
	// TODO: optimize this, https://stackoverflow.com/questions/8674718/best-way-to-select-random-rows-postgresql

	var projects []models.Project

	db := r.db
	db = db.Joins("LEFT JOIN regions ON regions.id = projects.region_id")
	for _, filter := range filters {
		db = filter.Apply(db)
	}

	err := db.
		Order("RANDOM()").
		Limit(n).
		Find(&projects).
		Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}
