package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino-ext/components/tool/googlesearch"
)

func TestDuckDuckGoSearch(t *testing.T) {
	ctx := context.Background()

	config := &duckduckgo.Config{
		MaxResults: 3,
		Region:     ddgsearch.RegionCN,
		DDGConfig: &ddgsearch.Config{
			Timeout:    30 * time.Second,
			Cache:      true,
			MaxRetries: 3,
		},
	}

	tool, err := duckduckgo.NewTool(ctx, config)
	if err != nil {
		log.Fatalf("NewTool of duckduckgo failed, err=%v", err)
	}

	searchReq := &duckduckgo.SearchRequest{
		Query: "Golang programming development",
		Page:  1,
	}

	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		log.Fatalf("Marshal of search request failed, err=%v", err)
	}

	resp, err := tool.InvokableRun(ctx, string(jsonReq))
	if err != nil {
		log.Fatalf("Search of duckduckgo failed, err=%v", err)
	}

	var searchResp duckduckgo.SearchResponse
	if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
		log.Fatalf("Unmarshal of search response failed, err=%v", err)
	}

	fmt.Println("Search Results:")
	for i, result := range searchResp.Results {
		fmt.Printf("%d. Title: %s\n", i+1, result.Title)
		fmt.Printf("		Link: %s\n", result.Link)
		fmt.Printf("		Description: %s\n", result.Description)
	}

}

func TestGoogleSearch(t *testing.T) {
	ctx := context.Background()

	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	googleSearchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if googleAPIKey == "" || googleSearchEngineID == "" {
		log.Fatal("[GOOGLE_API_KEY] and [GOOGLE_SEARCH_ENGINE_ID] must set")
	}

	// create tool
	searchTool, err := googlesearch.NewTool(ctx, &googlesearch.Config{
		APIKey:         googleAPIKey,
		SearchEngineID: googleSearchEngineID,
		Lang:           "zh-CN",
		Num:            5,
	})
	if err != nil {
		log.Fatal(err)
	}

	// prepare params
	req := googlesearch.SearchRequest{
		Query: "Golang concurrent programming",
		Num:   3,
		Lang:  "en",
	}

	args, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	// do search
	resp, err := searchTool.InvokableRun(ctx, string(args))
	if err != nil {
		log.Fatal(err)
	}

	var searchResp googlesearch.SearchResult
	if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
		log.Fatal(err)
	}

	// Print results
	fmt.Println("Search Results:")
	fmt.Println("==============")
	for i, result := range searchResp.Items {
		fmt.Printf("\n%d. Title: %s\n", i+1, result.Title)
		fmt.Printf("   Link: %s\n", result.Link)
		fmt.Printf("   Desc: %s\n", result.Desc)
	}
	fmt.Println("")
	fmt.Println("==============")
}
