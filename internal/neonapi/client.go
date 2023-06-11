package neonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
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

func (c *Client) sendJSONRequest(method string, path string, requestObj any, responseObj any) error {
	url := c.baseURL + path

	log.WithFields(log.Fields{
		"method":  method,
		"url":     url,
		"request": requestObj,
	}).Info("sending request")

	var reader io.Reader
	if requestObj != nil {
		requestBody, err := json.Marshal(requestObj)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(requestBody)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")

	// TODO: custom client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: don't store the whole body in memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"method": method,
		"url":    url,
		"status": resp.Status,
		"body":   json.RawMessage(body),
	}).Info("got response")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("got status code %d, body = %s", resp.StatusCode, body)
	}

	if responseObj != nil {
		err = json.Unmarshal(body, responseObj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CreateProject(req *CreateProject) (*CreateProjectResponse, error) {
	// https://api-docs.neon.tech/reference/createproject
	var resp CreateProjectResponse
	err := c.sendJSONRequest("POST", "/projects", &CreateProjectRequest{
		Project: req,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteProject(projectID string) error {
	// https://api-docs.neon.tech/reference/deleteproject
	return c.sendJSONRequest("DELETE", fmt.Sprintf("/projects/%s", projectID), nil, nil)
}
