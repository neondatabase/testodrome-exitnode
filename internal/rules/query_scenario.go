package rules

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/drivers"
	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
)

const av1Select1 = `SELECT 1`
const av1CreateTable = `CREATE TABLE IF NOT EXISTS activity_v1 (
		id SERIAL PRIMARY KEY,
		nonce BIGINT,
		val FLOAT,
		created_at TIMESTAMP DEFAULT NOW()
	  )`
const av1DoActivity = `INSERT INTO activity_v1(nonce,val) SELECT $1 AS nonce, avg(id) AS val FROM activity_v1 RETURNING *`

type queryParams struct {
	project *models.Project
	driver  drivers.Driver
}

type queryScenario interface {
	execute(ctx context.Context, params queryParams) error
	exclusive() bool
}

func getScenario(name string) (queryScenario, error) {
	switch name {
	case "activityV1":
		return &activityV1{}, nil
	case "alwaysOn":
		return &alwaysOn{}, nil
	case "awaitShutdown":
		return &awaitShutdown{}, nil
	}

	return nil, fmt.Errorf("unknown scenario: %v", name)
}

// activityV1 is a simple scenario that executes 3 queries.
type activityV1 struct{}

func (a *activityV1) exclusive() bool {
	return false
}

func (a *activityV1) execute(ctx context.Context, params queryParams) error {
	queries := []drivers.SingleQuery{
		// first query, can trigger a cold start
		{Query: av1Select1},
		// init table
		{Query: av1CreateTable},
		// do some activity
		{Query: av1DoActivity, Params: []any{rand.Int63()}},
	}

	return executeManyQueries(ctx, params.driver, queries)
}

// alwaysOn will execute queries in a loop.
type alwaysOn struct{}

func (a *alwaysOn) exclusive() bool {
	return false
}

func (a *alwaysOn) execute(ctx context.Context, params queryParams) error {
	suspendTimeout := params.project.SuspendTimeout()
	// we want to make one query at least every 1/4 of suspend timeout
	targetDuration := suspendTimeout / 4

	var err error

	// wake up
	err = executeManyQueries(ctx, params.driver, []drivers.SingleQuery{
		{Query: av1Select1},
	})
	if err != nil {
		return err
	}

	lastQuery := time.Now()
	// init the database
	err = executeManyQueries(ctx, params.driver, []drivers.SingleQuery{
		{Query: av1CreateTable},
	})
	if err != nil {
		return err
	}

	for {
		now := time.Now()
		timeSpent := now.Sub(lastQuery)
		if timeSpent > targetDuration {
			return fmt.Errorf("last query was too long ago: %s, suspend timeout is %s", now.Sub(lastQuery), suspendTimeout)
		}

		timeSpent -= targetDuration

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(targetDuration - timeSpent):
			// ok
		}

		lastQuery = time.Now()
		err = executeManyQueries(ctx, params.driver, []drivers.SingleQuery{
			{Query: av1DoActivity, Params: []any{rand.Int63()}},
		})
		if err != nil {
			return err
		}
	}
}

type awaitShutdown struct{}

func (a *awaitShutdown) exclusive() bool {
	// we don't want any other process to interfere with this one
	return true
}

func (a *awaitShutdown) execute(ctx context.Context, params queryParams) error {
	// wake up + init
	err := executeManyQueries(ctx, params.driver, []drivers.SingleQuery{
		{Query: av1Select1},
		{Query: av1CreateTable},
	})
	if err != nil {
		return err
	}

	// wait for database to shut down
	targetTime := params.project.SuspendTimeout()
	// add 10% to the target time, so we don't query before shutdown
	targetTime = targetTime + targetTime/10 + time.Second*10

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(targetTime):
		// ok
	}

	// wake up the database and do some activity
	err = executeManyQueries(ctx, params.driver, []drivers.SingleQuery{
		{Query: av1DoActivity, Params: []any{rand.Int63()}},
	})
	if err != nil {
		return err
	}

	return nil
}

func executeManyQueries(ctx context.Context, driver drivers.Driver, queries []drivers.SingleQuery) error {
	log.Info(ctx, "executing queries", zap.Int("count", len(queries)))

	var res []models.Query
	var err error
	if md, ok := driver.(drivers.ManyQueriesDriver); ok {
		res, err = md.Queries(ctx, queries...)
	} else {
		for i, q := range queries {
			ctx := ctx
			if i > 0 {
				ctx = drivers.NotColdContext(ctx)
			}

			var query *models.Query
			query, err = driver.Query(ctx, q)
			if query != nil {
				res = append(res, *query)
			}
			if err != nil {
				break
			}
		}
	}

	_ = res
	return err
}
