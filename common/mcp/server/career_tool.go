package mcp

import (
	"context"
	"math"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterCareerTools(mcpServer *server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool(
			"analyze_job_match",
			mcp.WithDescription("分析岗位 JD 和用户简历/项目经历的匹配度"),
			mcp.WithString("jd", mcp.Description("岗位 JD"), mcp.Required()),
			mcp.WithString("resume_context", mcp.Description("简历或项目经历上下文"), mcp.Required()),
			mcp.WithString("preference", mcp.Description("用户偏好，如城市/方向/技术栈")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			jd, err := GetStringArg(args, "jd", true)
			if err != nil {
				return nil, err
			}
			resume, err := GetStringArg(args, "resume_context", true)
			if err != nil {
				return nil, err
			}
			preference, _ := GetStringArg(args, "preference", false)
			score := 60 + int(math.Min(float64(len(commonWords(jd, resume))*8), 30))
			if preference != "" && strings.Contains(strings.ToLower(jd), strings.ToLower(preference)) {
				score += 5
			}
			if score > 95 {
				score = 95
			}
			return JSONText(map[string]any{
				"match_score": score,
				"matched_points": []string{
					"项目包含 Go/Gin 后端、会话管理和 SSE 流式输出。",
					"涉及 RAG、Redis、MySQL、RabbitMQ，符合 AI 应用工程化要求。",
					"GradPilot Agent 可作为岗位相关的 AI 项目亮点。",
				},
				"missing_points": []string{
					"真实线上压测数据、用户规模和成本指标需要补充。",
					"外部搜索目前是 mock，面试时要说明扩展边界。",
				},
				"suggestions": []string{
					"简历中突出 Tool Registry、Memory、Skill Loader 的职责边界。",
					"用量化表达描述 SSE 响应体验、异步落库和模块可扩展性。",
				},
			})
		},
	)

	mcpServer.AddTool(
		mcp.NewTool(
			"generate_interview_questions",
			mcp.WithDescription("根据项目和岗位生成面试追问"),
			mcp.WithString("project", mcp.Description("项目名称或描述"), mcp.Required()),
			mcp.WithString("jd", mcp.Description("岗位 JD")),
			mcp.WithString("difficulty", mcp.Description("easy / medium / hard")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()
			project, err := GetStringArg(args, "project", true)
			if err != nil {
				return nil, err
			}
			jd, _ := GetStringArg(args, "jd", false)
			difficulty, _ := GetStringArg(args, "difficulty", false)
			if difficulty == "" {
				difficulty = "medium"
			}
			return JSONText([]string{
				"[" + difficulty + "] 你为什么把 " + project + " 设计成 Agentic RAG，而不是普通 RAG？",
				"Skill、Tool、MCP、Memory、RAG 的边界分别是什么？",
				"如果 MCP 工具调用失败，系统如何降级并保持 SSE 响应稳定？",
				"PaperChunker 为什么要保留 section metadata？后续如何做 parent-child chunking？",
				"结合 JD：" + emptyFallback(jd, "未提供 JD") + "，你会如何把项目亮点改写成岗位相关表达？",
			})
		},
	)
}

func commonWords(a, b string) []string {
	words := map[string]bool{}
	for _, word := range strings.Fields(strings.ToLower(a)) {
		words[word] = true
	}
	var out []string
	for _, word := range strings.Fields(strings.ToLower(b)) {
		if words[word] {
			out = append(out, word)
		}
	}
	return out
}
