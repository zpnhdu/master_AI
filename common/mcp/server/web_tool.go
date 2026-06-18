package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterWebTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"fetch_web_page",
			mcp.WithDescription("抓取网页详情，当前返回 mock 网页内容"),
			mcp.WithString("url", mcp.Description("网页 URL"), mcp.Required()),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			url, err := GetStringArg(request.GetArguments(), "url", true)
			if err != nil {
				return nil, err
			}
			return JSONText(map[string]any{
				"title":   "Mock page for " + url,
				"content": "该页面摘要由 GradPilot mock 生成，可用于演示抓取后的二次总结流程。内容包含研究背景、关键方法、可复现资源和注意事项。",
				"url":     url,
			})
		},
	)
}
