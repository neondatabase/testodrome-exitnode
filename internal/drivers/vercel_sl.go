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

// websockets
const VercelEdge04 = "https://sl-driver.vercel.app/api/query"

// http
const VercelEdge07 = "https://sl-driver.vercel.app/api/v07/http_query"
const VercelEdge08 = "https://sl-driver.vercel.app/api/v08/http_query"
const VercelNode09 = "https://neon-vercel-node.vercel.app/api/query"
const VercelNode09WS = "https://neon-vercel-node.vercel.app/api/ws_query"

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

var _ Driver = (*VercelSL)(nil)
var _ ManyQueriesDriver = (*VercelSL)(nil)

// VercelSL is `@neondatabase/serverless` deployed on Vercel.
type VercelSL struct {
	saver   QuerySaver
	connstr string
	apiURL  string
}

func NewVercelSL(connstr string, saver QuerySaver, apiURL string) *VercelSL {
	return &VercelSL{
		connstr: connstr,
		apiURL:  apiURL,
		saver:   saver,
	}
}

func (s *VercelSL) Query(ctx context.Context, singleQuery SingleQuery) (*models.Query, error) {
	res, err := s.Queries(ctx, singleQuery)
	var q *models.Query
	if len(res) == 1 {
		q = &res[0]
	}
	return q, err
}

func (s *VercelSL) Queries(ctx context.Context, queries ...SingleQuery) ([]models.Query, error) {
	res, err := s.queries(ctx, queries...)

	for i := range res {
		if err2 := saveQuery(s.saver, &res[i], err); err2 != nil {
			return res, err2
		}
	}

	return res, err
}

func (s *VercelSL) queries(ctx context.Context, queries ...SingleQuery) ([]models.Query, error) {
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
	req.Header.Set("Content-Type", "application/json")

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

	log.Debug(ctx, "got response", zap.Any("body", json.RawMessage(body)))

	var slResp slResponse
	if err := json.Unmarshal(body, &slResp); err != nil {
		log.Error(ctx, "failed to unmarshal response", zap.String("body", string(body)))
		// save invalid response to the database

		failedQuery := startQuery(
			ctx,
			models.QueryDB,
			s.connstr,
			"vercel-sl",
			"POST",
			string(requestBody),
		)
		finishQuery(failedQuery, string(body), err)

		return nil, saveQuery(s.saver, failedQuery, err)
	}

	var retQueries []models.Query
	for _, slQuery := range slResp.Queries {
		retQueries = append(retQueries, s.convert(slQuery))
	}
	return retQueries, nil
}

func (s *VercelSL) convert(slQuery slQueryResponse) models.Query {
	return models.Query{
		Exitnode: slQuery.Exitnode,
		Kind:     models.QueryDestination(slQuery.Kind),
		Addr:     slQuery.Addr,
		Driver:   slQuery.Driver,
		Method:   slQuery.Method,
		Request:  slQuery.Request,
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
