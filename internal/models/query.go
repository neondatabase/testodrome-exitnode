package models

import "time"

type QueryDestination string

const (
	// Query to the database (SQL)
	QueryDB QueryDestination = "db"
	// Request to the HTTP API
	QueryAPI QueryDestination = "api"
)

type Query struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// For DB queries it's the ID of the queried project.
	ProjectID *uint

	// RegionID is a foreign key to the region.
	RegionID uint

	// The node that executed the query.
	Exitnode string

	// Usually refers to a connection establishment query.
	RelatedQueryID *uint

	// Query to "api" or "db".
	Kind QueryDestination

	// For API queries it's the full URL of the API endpoint.
	// For postgres it's the connection string (with password included).
	Addr string

	// Driver is the name of the driver that executed the query.
	// For SQL queries it can be "go-serverless" (SQL over HTTP), "pgx" (popular go driver for PostgreSQL), etc.
	Driver string

	// For API queries it can be "create_project", "delete_project", etc.
	// For database queries it can be "connect", "query-sql".
	Method string

	// What was sent to the node. SQL query or HTTP request body.
	Request string

	// Result is available only for finished queries.
	QueryResult

	// NotCold is true if the query is not the first in the chain.
	// That means that it's most likely not a cold start.
	NotCold bool
}

// QueryResult is available only for finished queries.
type QueryResult struct {
	// IsFinished is true if the query is fully finished, and no process
	// will update it in the future.
	IsFinished bool

	// What was received from the node. SQL response or HTTP response body.
	Response string
	// Error message if the query failed.
	Error string
	// Timestamp when the query was started.
	StartedAt *time.Time
	// Timestamp when the query was finished.
	FinishedAt *time.Time
	// IsFailed is true if the query is failed.
	IsFailed bool
	// Duration is the duration of the query.
	Duration *time.Duration
}

func QueryDuration(ns *int64) *time.Duration {
	if ns == nil {
		return nil
	}
	d := time.Duration(*ns)
	return &d
}
