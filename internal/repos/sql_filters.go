package repos

import (
	"gorm.io/gorm"
)

type Filter interface {
	Apply(query *gorm.DB) *gorm.DB
}

type WhereFilter struct {
	SQL  string
	Args []any
}

func (f WhereFilter) Apply(query *gorm.DB) *gorm.DB {
	return query.Where(f.SQL, f.Args...)
}

func FilterByRegionProvider(provider string) WhereFilter {
	return WhereFilter{
		SQL:  "regions.provider = ?",
		Args: []any{provider},
	}
}

func RawFilter(sql string) WhereFilter {
	return WhereFilter{
		SQL: sql,
	}
}
