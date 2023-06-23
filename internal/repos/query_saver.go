package repos

import "github.com/petuhovskiy/neon-lights/internal/models"

type QuerySaverArgs struct {
	ProjectID *uint
	RegionID  *uint
	Exitnode  *string
}

func (a *QuerySaverArgs) Apply(q *models.Query) {
	if q.ProjectID == nil {
		q.ProjectID = a.ProjectID
	}
	if q.RegionID == 0 && a.RegionID != nil {
		q.RegionID = *a.RegionID
	}
	if q.Exitnode == "" && a.Exitnode != nil {
		q.Exitnode = *a.Exitnode
	}
}

// QuerySaver modifies and saves queries.
type QuerySaver struct {
	repo *QueryRepo
	args QuerySaverArgs
}

func NewQuerySaver(repo *QueryRepo, args QuerySaverArgs) *QuerySaver {
	return &QuerySaver{
		repo: repo,
		args: args,
	}
}

func (s *QuerySaver) Save(query *models.Query) error {
	s.args.Apply(query)
	return s.repo.Save(query)
}
