package repos

import (
	"gorm.io/gorm"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

type RegionRepo struct {
	db *gorm.DB
}

func NewRegionRepo(db *gorm.DB) *RegionRepo {
	return &RegionRepo{
		db: db,
	}
}

// Find returns all region filtered by the given filters.
func (r *RegionRepo) Find(filters []Filter) ([]models.Region, error) {
	var regions []models.Region

	db := r.db
	for _, filter := range filters {
		db = filter.Apply(db)
	}

	err := db.Find(&regions).Error
	if err != nil {
		return nil, err
	}
	return regions, nil
}
