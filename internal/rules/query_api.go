package rules

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/neonapi"
	"github.com/petuhovskiy/neon-lights/internal/repos"
)

func queryAPI[T any](ctx context.Context, prep *neonapi.Prepared[T], saver *repos.QuerySaver) (*T, error) {
	dbQuery := prep.QueryNoArgs()
	err := saver.Save(dbQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to persist query: %w", err)
	}

	resp, result, err := prep.Do(ctx)
	dbErr := saver.FinishSaveResult(dbQuery, result)

	// 1. save response to the database
	if dbErr != nil {
		log.Error(ctx, "failed to persist query result", zap.Error(dbErr))
		if err == nil {
			err = dbErr
		} else {
			err = errors.Join(err, dbErr)
		}
	}

	// 2. return
	return resp, err
}
