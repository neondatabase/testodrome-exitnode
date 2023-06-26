package repos

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/petuhovskiy/neon-lights/internal/models"
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

func FilterByRegionID(id uint) WhereFilter {
	return WhereFilter{
		SQL:  "regions.id = ?",
		Args: []any{id},
	}
}

func RawFilter(sql string) WhereFilter {
	return WhereFilter{
		SQL: sql,
	}
}

type MatrixFilterer struct {
	obj    any
	fields []string
}

func (f *MatrixFilterer) Apply(query *gorm.DB) *gorm.DB {
	var args []any
	for _, field := range f.fields {
		args = append(args, field)
	}
	return query.Where(f.obj, args...)
}

func ProjectMatrixFilter(proj *models.Project, fields []string) *MatrixFilterer {
	return &MatrixFilterer{
		obj:    proj,
		fields: fields,
	}
}

func MatrixFilters(project *models.Project, fields []string) ([]Filter, error) {
	var filters []Filter
	var projectFields []string

	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "projects."):
			projectFields = append(projectFields, strings.TrimPrefix(field, "projects."))

		default:
			return nil, fmt.Errorf("unsupported field filter %q", field)
		}
	}

	if projectFields != nil {
		if project == nil {
			return nil, fmt.Errorf("project fields filter requires project")
		}
		filters = append(filters, ProjectMatrixFilter(project, projectFields))
	}

	return filters, nil
}
