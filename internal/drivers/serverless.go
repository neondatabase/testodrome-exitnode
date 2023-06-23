package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

type hosRequest struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

var _ Driver = (*Serverless)(nil)

// SQL-over-HTTP driver.
type Serverless struct {
	connstr *url.URL
	saver   QuerySaver
}

func NewServerless(connectionString string, saver QuerySaver) (*Serverless, error) {
	connstr, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}
	return &Serverless{
		connstr: connstr,
		saver:   saver,
	}, nil
}

func (s *Serverless) httpURL() string {
	return fmt.Sprintf("https://%s/sql", s.connstr.Hostname())
}

func (s *Serverless) Query(ctx context.Context, singleQuery SingleQuery) (*models.Query, error) {
	q, err := s.query(ctx, singleQuery)
	return q, saveQuery(s.saver, q, err)
}

func (s *Serverless) query(ctx context.Context, singleQuery SingleQuery) (retQuery *models.Query, retErr error) {
	query := singleQuery.Query
	params := singleQuery.Params
	if params == nil {
		params = []any{}
	}

	hos := hosRequest{
		Query:  query,
		Params: params,
	}
	requestBody, err := json.Marshal(hos)
	if err != nil {
		return nil, err
	}

	retQuery = startQuery(
		models.QueryDB,
		s.connstr.String(),
		"go-serverless",
		"sql-over-http",
		string(requestBody),
	)

	defer func() {
		if retErr != nil {
			retQuery.IsFinished = true
			retQuery.IsFailed = true
			retQuery.Error = retErr.Error()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", s.httpURL(), bytes.NewReader(requestBody))
	if err != nil {
		return retQuery, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Neon-Connection-String", s.connstr.String())

	startedAt := time.Now()
	retQuery.StartedAt = &startedAt

	// TODO: non-default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retQuery, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return retQuery, err
	}

	finishedAt := time.Now()
	duration := finishedAt.Sub(startedAt)
	retQuery.FinishedAt = &finishedAt
	retQuery.Duration = &duration
	retQuery.IsFinished = true
	retQuery.Response = string(body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return retQuery, fmt.Errorf("bad status code %d, body: %s", resp.StatusCode, body)
	}

	return retQuery, nil
}
