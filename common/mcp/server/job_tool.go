package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterJobTools(mcpServer *server.MCPServer) {
	RegisterJobToolsWithClient(mcpServer, NewMockJobSearchClient())
}

func RegisterJobToolsWithClient(mcpServer *server.MCPServer, client JobSearchClient) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_job_jd",
			mcp.WithDescription("搜索岗位 JD，默认使用 mock client，后续可替换搜索 API / Browser MCP / 招聘网站适配器"),
			mcp.WithString("keyword", mcp.Description("岗位关键词"), mcp.Required()),
			mcp.WithString("city", mcp.Description("可选城市")),
			mcp.WithNumber("limit", mcp.Description("返回数量，默认 3，最多 8")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			keyword, err := GetStringArg(args, "keyword", true)
			if err != nil {
				return nil, err
			}
			city, _ := GetStringArg(args, "city", false)
			if city == "" {
				city = "remote"
			}
			limit := clampLimit(GetIntArg(args, "limit", 3))
			jobs, err := client.SearchJobs(ctx, keyword, city, limit)
			if err != nil {
				return nil, err
			}
			return JSONText(jobs)
		},
	)
}
