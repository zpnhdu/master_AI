package mcp

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterResearchTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"generate_meeting_outline",
			mcp.WithDescription("生成科研组会汇报大纲"),
			mcp.WithString("topic", mcp.Description("汇报主题"), mcp.Required()),
			mcp.WithString("duration", mcp.Description("汇报时长，如 20 min")),
			mcp.WithString("paper_summary", mcp.Description("论文摘要或用户材料摘要")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			topic, err := GetStringArg(args, "topic", true)
			if err != nil {
				return nil, err
			}
			duration, _ := GetStringArg(args, "duration", false)
			if duration == "" {
				duration = "20 min"
			}
			paperSummary, _ := GetStringArg(args, "paper_summary", false)
			return JSONText(map[string]any{
				"background": "用 2-3 页说明 " + topic + " 的研究背景、应用场景和为什么值得做。",
				"problem":    "定义核心问题、现有方法不足和本次汇报要回答的问题。",
				"method":     "拆解模型/系统流程，突出关键模块、数据流和创新点。",
				"result":     "展示主要实验、指标、消融和与 baseline 的差异。",
				"discussion": "结合论文摘要或用户材料讨论局限：" + emptyFallback(paperSummary, "数据规模、泛化性、复现成本和工程落地风险。"),
				"next_plan":  "按 " + duration + " 汇报节奏给出 2-3 个下一步实验或工程验证。",
			})
		},
	)

	mcpServer.AddTool(
		mcp.NewTool(
			"extract_keywords",
			mcp.WithDescription("从论文、JD、简历、项目描述中抽关键词"),
			mcp.WithString("text", mcp.Description("待抽取文本"), mcp.Required()),
			mcp.WithString("domain", mcp.Description("research / career / general")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			text, err := GetStringArg(args, "text", true)
			if err != nil {
				return nil, err
			}
			domain, _ := GetStringArg(args, "domain", false)
			keywords := extractKeywords(text, domain)
			return JSONText(map[string]any{"keywords": keywords})
		},
	)
}

func extractKeywords(text, domain string) []string {
	base := []string{}
	seen := map[string]bool{}
	for _, field := range strings.Fields(strings.ToLower(text)) {
		field = strings.Trim(field, " ,.;:，。；：!?！？()[]{}")
		if len([]rune(field)) < 3 || seen[field] {
			continue
		}
		seen[field] = true
		base = append(base, field)
		if len(base) >= 8 {
			break
		}
	}
	switch domain {
	case "research":
		base = append(base, "method", "experiment", "limitation")
	case "career":
		base = append(base, "jd", "resume", "impact")
	default:
		base = append(base, "context", "objective")
	}
	return base
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
