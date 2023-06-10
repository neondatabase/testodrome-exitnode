package repos

import (
	"fmt"

	"gorm.io/gorm"
)

type SequenceRepo struct {
	db *gorm.DB
}

func NewSequenceRepo(db *gorm.DB) *SequenceRepo {
	return &SequenceRepo{
		db: db,
	}
}

// Get returns the sequence with the given key. If it does not exist, it is created.
func (r *SequenceRepo) Get(key string) (*Sequence, error) {
	// insert if not exists
	err := r.db.
		Exec("INSERT INTO sequences (key, val) VALUES (?, 0) ON CONFLICT DO NOTHING", key).
		Error
	if err != nil {
		return nil, fmt.Errorf("create sequence: %w", err)
	}

	return &Sequence{
		db:  r.db,
		key: key,
	}, nil
}

type Sequence struct {
	db  *gorm.DB
	key string
}

func (s *Sequence) Next() (uint, error) {
	var val uint
	err := s.db.
		Raw("UPDATE sequences SET val = val + 1 WHERE key = ? RETURNING val", s.key).
		Scan(&val).
		Error
	if err != nil {
		return 0, fmt.Errorf("update sequence: %w", err)
	}
	return val, nil
}
