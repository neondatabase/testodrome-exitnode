package neonapi

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	// TODO: this is a quick hack to enable JSON logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func testClient(t *testing.T) *Client {
	apiKey := os.Getenv("NEON_API_KEY")
	if apiKey == "" {
		t.Skip("NEON_API_KEY is not set")
	}
	return NewClient("stage.neon.tech", apiKey)
}

// Run with `export $(cat .env | xargs) && go test ./... -v -run TestCreateProject`
func TestCreateProject(t *testing.T) {
	client := testClient(t)
	resp, err := client.CreateProject(&CreateProject{
		Name:     "test-project-1",
		RegionID: "aws-eu-west-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Project ID: %s", resp.Project.ID)
}
