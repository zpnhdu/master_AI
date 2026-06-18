package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPaperTools(mcpServer *server.MCPServer) {
	RegisterPaperToolsWithClient(mcpServer, NewMockPaperSearchClient())
}

func RegisterPaperToolsWithClient(mcpServer *server.MCPServer, client PaperSearchClient) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_papers",
			mcp.WithDescription("搜索科研论文，默认使用 mock client，后续可替换 arXiv / Semantic Scholar"),
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
			papers, err := client.SearchPapers(ctx, query, limit)
			if err != nil {
				return nil, err
			}
			return JSONText(papers)
		},
	)
}
