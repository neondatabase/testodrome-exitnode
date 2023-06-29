package drivers

// Single query to the driver.
type SingleQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

type Name string

const (
	PgxConn      = Name("pgx-conn")
	GoServerless = Name("go-serverless")
	VercelEdge   = Name("vercel-edge")
)
