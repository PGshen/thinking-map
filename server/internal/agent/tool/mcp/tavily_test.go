package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGetTavilyTools_NoAPIKey(t *testing.T) {
	// 备份并清理环境变量
	orig := os.Getenv("TAVILY_API_KEY")
	os.Unsetenv("TAVILY_API_KEY")
	t.Cleanup(func() {
		if orig != "" {
			_ = os.Setenv("TAVILY_API_KEY", orig)
		}
	})

	ctx := context.Background()
	_, err := GetTavilyTools(ctx, nil)
	if err == nil {
		t.Fatalf("expected error when API key missing")
	}
}

func TestBuildTavilyURL(t *testing.T) {
	apiKey := "tvly-TEST"
	cases := []string{
		"https://mcp.tavily.com/mcp/",
		"https://mcp.tavily.com/mcp",
		"https://mcp.tavily.com/mcp/?",
		"https://mcp.tavily.com/mcp?existing=1",
	}
	for _, base := range cases {
		got := buildTavilyURL(base, apiKey)
		if got == "" {
			t.Fatalf("empty url for base %s", base)
		}
		if base == "https://mcp.tavily.com/mcp?existing=1" {
			if got != "https://mcp.tavily.com/mcp?existing=1&tavilyApiKey="+apiKey {
				t.Fatalf("unexpected url: %s", got)
			}
		} else {
			if !strings.Contains(got, "tavilyApiKey="+apiKey) {
				t.Fatalf("api key missing in url: %s", got)
			}
		}
	}
}

// 可选集成测试（需要有效的 TAVILY_API_KEY，且网络可访问远程 MCP Server）
func TestGetTavilyTools_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}
	os.Setenv("TAVILY_API_KEY", "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")

	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		t.Skip("TAVILY_API_KEY not set, skip integration test")
	}
	ctx := context.Background()
	tools, err := GetTavilyTools(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get tavily tools: %v", err)
	}
	if len(tools) == 0 {
		t.Fatalf("tavily tools should not be empty")
	}
	strTools, _ := json.Marshal(tools)
	fmt.Printf("%s\n", string(strTools))
}
