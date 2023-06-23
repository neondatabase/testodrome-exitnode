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

// SQL-over-HTTP driver.
type Serverless struct {
	projectID uint
	regionID  uint
	exitnode  string
	connstr   *url.URL
}

func NewServerless(exitnode string, project *models.Project) (*Serverless, error) {
	connstr, err := url.Parse(project.ConnectionString)
	if err != nil {
		return nil, err
	}
	return &Serverless{
		projectID: project.ID,
		regionID:  project.RegionID,
		exitnode:  exitnode,
		connstr:   connstr,
	}, nil
}

func (s *Serverless) httpURL() string {
	return fmt.Sprintf("https://%s/sql", s.connstr.Hostname())
}

func (s *Serverless) Query(ctx context.Context, query string, params ...any) (retQuery *models.Query, retErr error) {
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

	retQuery = &models.Query{
		ProjectID:   &s.projectID,
		RegionID:    s.regionID,
		Exitnode:    s.exitnode,
		Kind:        models.QueryDB,
		Addr:        s.connstr.String(),
		Driver:      "go-serverless",
		Method:      "sql-over-http",
		Request:     string(requestBody),
		QueryResult: models.QueryResult{},
	}

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
