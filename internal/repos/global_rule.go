package repos

import (
	"fmt"

	"github.com/petuhovskiy/neon-lights/internal/models"
	"gorm.io/gorm"
)

type GlobalRuleRepo struct {
	db *gorm.DB
}

func NewGlobalRuleRepo(db *gorm.DB) *GlobalRuleRepo {
	return &GlobalRuleRepo{
		db: db,
	}
}

func (r *GlobalRuleRepo) AllEnabled() ([]models.GlobalRule, error) {
	var rules []models.GlobalRule
	err := r.db.
		Where("enabled = ?", true).
		Order("priority ASC").
		Find(&rules).
		Error
	if err != nil {
		return nil, fmt.Errorf("find global rules: %w", err)
	}
	return rules, nil
}
