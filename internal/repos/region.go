package repos

import (
	"github.com/petuhovskiy/neon-lights/internal/models"
	"gorm.io/gorm"
)

type RegionRepo struct {
	db *gorm.DB
}

func NewRegionRepo(db *gorm.DB) *RegionRepo {
	return &RegionRepo{
		db: db,
	}
}

// FindByProvider returns all regions with the given provider.
func (r *RegionRepo) FindByProvider(providerName string) ([]models.Region, error) {
	var regions []models.Region
	err := r.db.
		Where("provider = ?", providerName).
		Find(&regions).
		Error
	if err != nil {
		return nil, err
	}
	return regions, nil
}
