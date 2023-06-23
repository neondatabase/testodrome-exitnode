package neonapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
)

// TODO: consider using https://github.com/kislerdm/neon-sdk-go instead

type Client struct {
	baseURL    string
	authHeader string
}

func NewClient(domain string, apiKey string) *Client {
	return &Client{
		baseURL:    fmt.Sprintf("https://console.%s/api/v2", domain),
		authHeader: fmt.Sprintf("Bearer %s", apiKey),
	}
}

func (c *Client) CreateProject(req *CreateProject) (*Prepared[CreateProjectResponse], error) {
	// https://api-docs.neon.tech/reference/createproject
	return prepare[CreateProjectResponse](c, "CreateProject", "POST", "/projects", &CreateProjectRequest{
		Project: req,
	})
}

func (c *Client) DeleteProject(projectID string) (*Prepared[DeleteProjectResponse], error) {
	// https://api-docs.neon.tech/reference/deleteproject
	return prepare[DeleteProjectResponse](c, "DeleteProject", "DELETE", fmt.Sprintf("/projects/%s", projectID), nil)
}

type Prepared[T any] struct {
	cli       *Client
	method    string
	fullURL   string
	body      json.RawMessage
	apiMethod string
}

func prepare[T any](cli *Client, apiMethod string, method string, path string, requestObj any) (*Prepared[T], error) {
	var body []byte
	if requestObj != nil {
		var err error
		body, err = json.Marshal(requestObj)
		if err != nil {
			return nil, err
		}
	}

	return &Prepared[T]{
		cli:       cli,
		method:    method,
		fullURL:   cli.baseURL + path,
		body:      body,
		apiMethod: apiMethod,
	}, nil
}

func (p *Prepared[T]) Query(projectID *uint, regionID uint, exitnode string) *models.Query {
	return &models.Query{
		ProjectID:   projectID,
		RegionID:    regionID,
		Exitnode:    exitnode,
		Kind:        models.QueryAPI,
		Addr:        p.fullURL,
		Driver:      "go-neonapi",
		Method:      p.apiMethod,
		Request:     string(p.body),
		QueryResult: models.QueryResult{},
	}
}

func (p *Prepared[T]) Do(ctx context.Context) (*T, *models.QueryResult, error) {
	result := &models.QueryResult{
		IsFinished: true,
		Response:   "",
		Error:      "",
		StartedAt:  &time.Time{},
		FinishedAt: &time.Time{},
		IsFailed:   false,
		Duration:   nil,
	}

	var responseObj T
	err := p.do(ctx, &responseObj, result)
	if err != nil {
		result.Error = err.Error()
		result.IsFailed = true
		return nil, result, err
	}

	return &responseObj, result, nil
}

func (p *Prepared[T]) do(ctx context.Context, responseObj any, result *models.QueryResult) error {
	ctx = log.With(
		ctx,
		zap.String("method", p.method),
		zap.String("url", p.fullURL),
	)

	log.Info(ctx, "sending request", zap.Any("request", p.body))

	var reader io.Reader
	if p.body != nil {
		reader = bytes.NewReader(p.body)
	}

	// TODO: set context
	req, err := http.NewRequestWithContext(context.Background(), p.method, p.fullURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.cli.authHeader)
	req.Header.Set("Accept", "application/json")

	startedAt := time.Now()
	result.StartedAt = &startedAt

	// TODO: custom client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	finishedAt := time.Now()
	duration := finishedAt.Sub(startedAt)
	result.FinishedAt = &finishedAt
	result.Duration = &duration
	result.Response = string(body)

	log.Info(
		ctx,
		"got response",
		zap.String("status", resp.Status),
		zap.Any("body", json.RawMessage(body)),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("got status code %d, body = %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(body, responseObj)
	if err != nil {
		return err
	}

	return nil
}
