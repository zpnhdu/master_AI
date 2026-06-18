package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterJobTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"search_job_jd",
			mcp.WithDescription("搜索岗位 JD，当前返回 mock 岗位列表"),
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
			jobs := make([]map[string]any, 0, limit)
			for i := 0; i < limit; i++ {
				jobs = append(jobs, map[string]any{
					"company":      fmt.Sprintf("AI Lab %d", i+1),
					"title":        fmt.Sprintf("%s Engineer", strings.Title(keyword)),
					"city":         city,
					"requirements": []string{"Go/Gin 后端经验", "RAG 或 LLM 应用经验", "Redis/MySQL/RabbitMQ 工程实践", "能清楚讲项目架构与优化"},
					"url":          fmt.Sprintf("https://jobs.example.com/%s/%d", slug(keyword), i+1),
				})
			}
			return JSONText(jobs)
		},
	)
}
