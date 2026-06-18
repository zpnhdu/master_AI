package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterGitHubTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_github_repos",
			mcp.WithDescription("搜索 GitHub 代码仓库，当前返回 mock 仓库列表"),
			mcp.WithString("query", mcp.Description("仓库搜索关键词"), mcp.Required()),
			mcp.WithString("language", mcp.Description("可选语言，如 Go、Python")),
			mcp.WithNumber("limit", mcp.Description("返回数量，默认 3，最多 8")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			query, err := GetStringArg(args, "query", true)
			if err != nil {
				return nil, err
			}
			language, _ := GetStringArg(args, "language", false)
			if language == "" {
				language = "Go"
			}
			limit := clampLimit(GetIntArg(args, "limit", 3))
			repos := make([]map[string]any, 0, limit)
			for i := 0; i < limit; i++ {
				repos = append(repos, map[string]any{
					"repo":        fmt.Sprintf("gradpilot/%s-%s-%d", slug(query), strings.ToLower(language), i+1),
					"description": fmt.Sprintf("%s implementation for %s with README, dataset notes, and reproducible scripts.", language, query),
					"stars":       320 + i*137,
					"language":    language,
					"url":         fmt.Sprintf("https://github.com/gradpilot/%s-%d", slug(query), i+1),
				})
			}
			return JSONText(repos)
		},
	)
}
