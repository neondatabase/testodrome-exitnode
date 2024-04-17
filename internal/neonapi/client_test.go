package neonapi

import (
	"context"
	"os"
	"testing"

	"github.com/petuhovskiy/neon-lights/internal/log"
)

func testClient(t *testing.T) *Client {
	apiKey := os.Getenv("NEON_API_KEY")
	if apiKey == "" {
		t.Skip("NEON_API_KEY is not set")
	}
	return NewClient("console-stage.neon.build", apiKey)
}

// Run with `export $(cat .env | xargs) && go test ./... -v -run TestCreateProject`
func TestCreateProject(t *testing.T) {
	_ = log.DefaultGlobals()
	ctx := context.Background()

	client := testClient(t)
	prep, err := client.CreateProject(&CreateProject{
		Name:        "test-project-1",
		RegionID:    "aws-eu-west-1",
		Branch:      CreateProjectBranch{RoleName: "testodrome"},
		PgVersion:   16,
		Provisioner: "k8s-neonvm",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, result, err := prep.Do(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(resp.Roles))
	}
	if resp.Roles[0].Name != "testodrome" {
		t.Fatalf("Expected role name 'testodrome', got %s", resp.Roles[0].Name)
	}

	t.Logf("Project ID: %s", resp.Project.ID)
	t.Logf("Result: %#+v", result)
}
