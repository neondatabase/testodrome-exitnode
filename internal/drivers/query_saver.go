package drivers

import (
	"context"
	"errors"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

// Hydrates query with additional data and saves it.
// These are the fields that can be updated:
// - ID, CreatedAt, UpdatedAt
// - ProjectID
// - RegionID
// - Exitnode
type QuerySaver interface {
	Save(query *models.Query) error
}

// Finishes the query and saves it by calling QuerySaver.Save.
// Returns the combined error from saver and queryErr.
func saveQuery(saver QuerySaver, query *models.Query, queryErr error) (retErr error) {
	retErr = queryErr
	if err := saver.Save(query); err != nil {
		if retErr == nil {
			retErr = err
		} else {
			retErr = errors.Join(retErr, err)
		}
	}

	return retErr
}

//nolint:unparam
func startQuery(
	ctx context.Context,
	kind models.QueryDestination,
	addr string,
	driver string,
	method string,
	request string,
) *models.Query {
	now := time.Now()
	return &models.Query{
		Kind:    kind,
		Addr:    addr,
		Driver:  driver,
		Method:  method,
		Request: request,
		QueryResult: models.QueryResult{
			StartedAt: &now,
		},
		NotCold: IsNotCold(ctx),
	}
}

func finishQuery(query *models.Query, response string, err error) {
	if query.Response == "" {
		query.Response = response
	}

	if err != nil && !query.IsFailed {
		query.IsFailed = true
		query.Error = err.Error()
	}

	query.IsFinished = true
	if query.FinishedAt == nil && query.StartedAt != nil {
		now := time.Now()
		query.FinishedAt = &now
	}

	if query.Duration == nil && query.StartedAt != nil && query.FinishedAt != nil {
		duration := query.FinishedAt.Sub(*query.StartedAt)
		query.Duration = &duration
	}
}
