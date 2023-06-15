package repos

import (
	"gorm.io/gorm"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

type QueryRepo struct {
	db *gorm.DB
}

func NewQueryRepo(db *gorm.DB) *QueryRepo {
	return &QueryRepo{
		db: db,
	}
}

// Save query to the database.
func (r *QueryRepo) Save(query *models.Query) error {
	return r.db.Save(query).Error
}

// Update result fields after the query was finished.
func (r *QueryRepo) FinishSaveResult(query *models.Query, upd *models.QueryResult) error {
	return r.db.Model(query).Updates(upd).Error
}
