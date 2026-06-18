package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterGitHubTools(mcpServer *server.MCPServer) {
	RegisterGitHubToolsWithClient(mcpServer, NewMockGitHubSearchClient())
}

func RegisterGitHubToolsWithClient(mcpServer *server.MCPServer, client GitHubSearchClient) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_github_repos",
			mcp.WithDescription("搜索 GitHub 代码仓库，默认使用 mock client，后续可替换 GitHub REST API"),
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
			repos, err := client.SearchRepos(ctx, query, language, limit)
			if err != nil {
				return nil, err
			}
			return JSONText(repos)
		},
	)
}
