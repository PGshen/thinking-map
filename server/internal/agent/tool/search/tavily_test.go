package search

import (
	"context"
	"os"
	"testing"

	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	global.SetupTestEnvironment()

	// run tests
	code := m.Run()

	// teardown

	os.Exit(code)
}

func TestNewTavilyClient(t *testing.T) {
	client := NewTavilyClient()
	assert.NotNil(t, client)
	assert.NotEmpty(t, client.APIKey)
	assert.NotNil(t, client.HttpClient)
}

func TestSearchFunc(t *testing.T) {
	if os.Getenv("TAVILY_API_KEY") == "" {
		t.Skip("TAVILY_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "mapID", "test-map-id")
	ctx = context.WithValue(ctx, "nodeID", "test-node-id")

	req := &SearchFuncRequest{
		Query: "who is newton?",
	}

	resp, err := SearchFunc(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Results)
}

func TestExtractFunc(t *testing.T) {
	if os.Getenv("TAVILY_API_KEY") == "" {
		t.Skip("TAVILY_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "mapID", "test-map-id")
	ctx = context.WithValue(ctx, "nodeID", "test-node-id")

	req := &ExtractFuncRequest{
		Query: "who is pgshen",
		URL:   "https://github.com/PGshen",
	}

	resp, err := ExtractFunc(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}
