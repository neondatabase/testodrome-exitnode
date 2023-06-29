package drivers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

const pgxDriverName = "go-pgx-conn"

var _ Driver = (*PgxConnection)(nil)

type PgxConnection struct {
	connstr   string
	conn      *pgx.Conn
	connQuery *models.Query
	saver     QuerySaver
}

func PgxConnect(ctx context.Context, connstr string, saver QuerySaver) (*PgxConnection, error) {
	connQuery := startQuery(
		ctx,
		models.QueryDB,
		connstr,
		pgxDriverName,
		"connect",
		"",
	)
	conn, err1 := pgx.Connect(ctx, connstr)
	finishQuery(connQuery, "", err1)

	if err := saveQuery(saver, connQuery, err1); err != nil {
		return nil, err
	}

	return &PgxConnection{
		connstr:   connstr,
		conn:      conn,
		connQuery: connQuery,
		saver:     saver,
	}, nil
}

func (c *PgxConnection) Query(ctx context.Context, req SingleQuery) (*models.Query, error) {
	jsonReq, err3 := json.Marshal(req)
	if err3 != nil {
		return nil, err3
	}

	query := startQuery(
		ctx,
		models.QueryDB,
		c.connstr,
		pgxDriverName,
		"query",
		string(jsonReq),
	)
	query.RelatedQueryID = &c.connQuery.ID

	rows, err1 := c.conn.Query(ctx, req.Query, req.Params...)
	// rows are always non-nil

	jsonArr, err2 := pgx.CollectRows(rows, func(row pgx.CollectableRow) (json.RawMessage, error) {
		values, err := row.Values()
		if err != nil {
			return nil, err
		}
		j, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		return j, nil
	})
	jsonRows, _ := json.Marshal(jsonArr)

	if err2 != nil {
		if err1 != nil {
			err2 = errors.Join(err1, err2)
		}

		finishQuery(query, string(jsonRows), err2)
		return query, saveQuery(c.saver, query, err2)
	}

	finishQuery(query, string(jsonRows), nil)
	return query, saveQuery(c.saver, query, nil)
}

func (c *PgxConnection) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}
