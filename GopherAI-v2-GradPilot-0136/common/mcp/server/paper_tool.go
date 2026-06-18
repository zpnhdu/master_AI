package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPaperTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_papers",
			mcp.WithDescription("搜索科研论文，当前返回与 query 相关的 mock 论文列表"),
			mcp.WithString("query", mcp.Description("论文搜索关键词"), mcp.Required()),
			mcp.WithNumber("limit", mcp.Description("返回数量，默认 3，最多 8")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			query, err := GetStringArg(args, "query", true)
			if err != nil {
				return nil, err
			}
			limit := clampLimit(GetIntArg(args, "limit", 3))
			topics := []string{"Agentic RAG", "tool-augmented retrieval", "scientific workflow planning", "LLM memory", "research assistant"}
			papers := make([]map[string]any, 0, limit)
			for i := 0; i < limit; i++ {
				topic := topics[i%len(topics)]
				papers = append(papers, map[string]any{
					"title":    fmt.Sprintf("%s for %s: A Practical Study", strings.Title(topic), query),
					"authors":  []string{"Chen Li", "Yuki Tanaka", "Maria Smith"},
					"year":     2023 + i%3,
					"abstract": fmt.Sprintf("This mock paper studies %s in the context of %s, focusing on retrieval planning, evidence grounding, and reproducible evaluation.", topic, query),
					"url":      fmt.Sprintf("https://example.org/papers/%s/%d", slug(query), i+1),
					"source":   "mock-arxiv",
				})
			}
			return JSONText(papers)
		},
	)
}
