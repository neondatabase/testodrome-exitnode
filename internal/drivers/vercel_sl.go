package drivers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
)

const defaultAPIURL = "https://sl-driver.vercel.app/api/query"

type slRequest struct {
	ConnStr string        `json:"connstr"`
	Queries []SingleQuery `json:"queries"`
}

type slResponse struct {
	Queries []slQueryResponse `json:"queries"`
}

type slQueryResponse struct {
	Exitnode   string     `json:"exitnode"`
	Kind       string     `json:"kind"`
	Addr       string     `json:"addr"`
	Driver     string     `json:"driver"`
	Method     string     `json:"method"`
	Request    string     `json:"request"`
	Response   string     `json:"response"`
	Error      string     `json:"error"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	IsFailed   bool       `json:"isFailed"`
	DurationNs *int64     `json:"durationNs"`
}

// Single query to the driver.
type SingleQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

// VercelSL is `@neondatabase/serverless` deployed on Vercel.
type VercelSL struct {
	projectID uint
	regionID  uint
	connstr   string
	apiURL    string
}

func NewVercelSL(project *models.Project) *VercelSL {
	return &VercelSL{
		projectID: project.ID,
		regionID:  project.RegionID,
		connstr:   project.ConnectionString,
		apiURL:    defaultAPIURL,
	}
}

func (s *VercelSL) Queries(ctx context.Context, queries ...SingleQuery) ([]models.Query, error) {
	slReq := slRequest{
		ConnStr: s.connstr,
		Queries: queries,
	}
	requestBody, err := json.Marshal(slReq)
	if err != nil {
		return nil, err
	}

	// TODO: set timeout or use provided context
	req, err := http.NewRequestWithContext(context.Background(), "POST", s.apiURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	// TODO: non-default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "got response", zap.Any("body", json.RawMessage(body)))

	var slResp slResponse
	if err := json.Unmarshal(body, &slResp); err != nil {
		return nil, err
	}

	var retQueries []models.Query
	for _, slQuery := range slResp.Queries {
		retQueries = append(retQueries, s.convert(slQuery))
	}
	return retQueries, nil
}

func (s *VercelSL) convert(slQuery slQueryResponse) models.Query {
	return models.Query{
		ProjectID: &s.projectID,
		RegionID:  s.regionID,
		Exitnode:  slQuery.Exitnode,
		Kind:      models.QueryDestination(slQuery.Kind),
		Addr:      slQuery.Addr,
		Driver:    slQuery.Driver,
		Method:    slQuery.Method,
		Request:   slQuery.Request,
		QueryResult: models.QueryResult{
			IsFinished: true,
			Response:   slQuery.Response,
			Error:      slQuery.Error,
			StartedAt:  slQuery.StartedAt,
			FinishedAt: slQuery.FinishedAt,
			IsFailed:   slQuery.IsFailed,
			Duration:   models.QueryDuration(slQuery.DurationNs),
		},
	}
}
