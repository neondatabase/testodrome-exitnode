package drivers

type Name string

const (
	PgxConn      = Name("pgx-conn")
	GoServerless = Name("go-serverless")
	VercelEdge   = Name("vercel-edge")
)
