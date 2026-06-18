package tools

import (
	"context"
	"fmt"
	"math"
	"strings"
)

func MCPToolInfos() []ToolInfo {
	return []ToolInfo{
		{Name: "get_weather", Description: "获取指定城市的天气信息。"},
		{Name: "search_papers", Description: "搜索科研论文 mock 列表。"},
		{Name: "search_github_repos", Description: "搜索 GitHub 代码仓库 mock 列表。"},
		{Name: "search_job_jd", Description: "搜索岗位 JD mock 列表。"},
		{Name: "fetch_web_page", Description: "抓取网页详情 mock 内容。"},
		{Name: "generate_meeting_outline", Description: "生成科研组会汇报大纲。"},
		{Name: "analyze_job_match", Description: "分析岗位 JD 与简历/项目经历匹配度。"},
		{Name: "generate_interview_questions", Description: "生成项目与岗位面试追问。"},
		{Name: "extract_keywords", Description: "从论文、JD、简历或项目描述中抽关键词。"},
	}
}

func CallBuiltinMCPTool(ctx context.Context, toolName string, args map[string]any) (string, error) {
	switch toolName {
	case "get_weather":
		city, err := getStringArg(args, "city", true)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("城市: %s\n温度: 23.5°C\n天气: 多云\n湿度: 58%%\n风速: 12.0 km/h\n说明: GradPilot mock weather，legacy MCP demo 可替换真实 wttr.in。", city), nil
	case "search_papers":
		query, err := getStringArg(args, "query", true)
		if err != nil {
			return "", err
		}
		limit := getIntArg(args, "limit", 3)
		return marshalJSON(mockPapers(query, limit))
	case "search_github_repos":
		query, err := getStringArg(args, "query", true)
		if err != nil {
			return "", err
		}
		language, _ := getStringArg(args, "language", false)
		limit := getIntArg(args, "limit", 3)
		return marshalJSON(mockRepos(query, language, limit))
	case "search_job_jd":
		keyword, err := getStringArg(args, "keyword", true)
		if err != nil {
			return "", err
		}
		city, _ := getStringArg(args, "city", false)
		limit := getIntArg(args, "limit", 3)
		return marshalJSON(mockJobs(keyword, city, limit))
	case "fetch_web_page":
		url, err := getStringArg(args, "url", true)
		if err != nil {
			return "", err
		}
		return marshalJSON(map[string]any{
			"title":   "Mock page for " + url,
			"content": "该页面摘要由 GradPilot mock 生成，可用于演示抓取后的二次总结流程。内容包含研究背景、关键方法、可复现资源和注意事项。",
			"url":     url,
		})
	case "generate_meeting_outline":
		topic, err := getStringArg(args, "topic", true)
		if err != nil {
			return "", err
		}
		duration, _ := getStringArg(args, "duration", false)
		summary, _ := getStringArg(args, "paper_summary", false)
		return marshalJSON(meetingOutline(topic, duration, summary))
	case "analyze_job_match":
		jd, err := getStringArg(args, "jd", true)
		if err != nil {
			return "", err
		}
		resume, err := getStringArg(args, "resume_context", true)
		if err != nil {
			return "", err
		}
		preference, _ := getStringArg(args, "preference", false)
		return marshalJSON(jobMatch(jd, resume, preference))
	case "generate_interview_questions":
		project, err := getStringArg(args, "project", true)
		if err != nil {
			return "", err
		}
		jd, _ := getStringArg(args, "jd", false)
		difficulty, _ := getStringArg(args, "difficulty", false)
		return marshalJSON(interviewQuestions(project, jd, difficulty))
	case "extract_keywords":
		text, err := getStringArg(args, "text", true)
		if err != nil {
			return "", err
		}
		domain, _ := getStringArg(args, "domain", false)
		return marshalJSON(map[string]any{"keywords": extractKeywords(text, domain)})
	default:
		return "", fmt.Errorf("unknown MCP tool: %s", toolName)
	}
}

func getIntArg(args map[string]any, key string, defaultValue int) int {
	value, ok := args[key]
	if !ok {
		return defaultValue
	}
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return defaultValue
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 3
	}
	if limit > 8 {
		return 8
	}
	return limit
}

func mockPapers(query string, limit int) []map[string]any {
	limit = clampLimit(limit)
	topics := []string{"Agentic RAG", "tool-augmented retrieval", "graduate research assistant", "scientific workflow planning", "LLM memory"}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		topic := topics[i%len(topics)]
		out = append(out, map[string]any{
			"title":    fmt.Sprintf("%s for %s: A Practical Study", strings.Title(topic), query),
			"authors":  []string{"Chen Li", "Yuki Tanaka", "Maria Smith"},
			"year":     2023 + i%3,
			"abstract": fmt.Sprintf("This mock paper studies %s in the context of %s, focusing on retrieval planning, evidence grounding, and reproducible evaluation.", topic, query),
			"url":      fmt.Sprintf("https://example.org/papers/%s/%d", slug(query), i+1),
			"source":   "mock-arxiv",
		})
	}
	return out
}

func mockRepos(query, language string, limit int) []map[string]any {
	limit = clampLimit(limit)
	if language == "" {
		language = "Go"
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, map[string]any{
			"repo":        fmt.Sprintf("gradpilot/%s-%s-%d", slug(query), strings.ToLower(language), i+1),
			"description": fmt.Sprintf("%s implementation for %s with README, dataset notes, and reproducible scripts.", language, query),
			"stars":       320 + i*137,
			"language":    language,
			"url":         fmt.Sprintf("https://github.com/gradpilot/%s-%d", slug(query), i+1),
		})
	}
	return out
}

func mockJobs(keyword, city string, limit int) []map[string]any {
	limit = clampLimit(limit)
	if city == "" {
		city = "remote"
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, map[string]any{
			"company":      fmt.Sprintf("AI Lab %d", i+1),
			"title":        fmt.Sprintf("%s Engineer", strings.Title(keyword)),
			"city":         city,
			"requirements": []string{"Go/Gin 后端经验", "RAG 或 LLM 应用经验", "Redis/MySQL/RabbitMQ 工程实践", "能清楚讲项目架构与优化"},
			"url":          fmt.Sprintf("https://jobs.example.com/%s/%d", slug(keyword), i+1),
		})
	}
	return out
}

func meetingOutline(topic, duration, paperSummary string) map[string]any {
	if duration == "" {
		duration = "20 min"
	}
	return map[string]any{
		"background": fmt.Sprintf("用 2-3 页说明 %s 的研究背景、应用场景和为什么值得做。", topic),
		"problem":    "定义核心问题、现有方法不足和本次汇报要回答的问题。",
		"method":     "拆解模型/系统流程，突出关键模块、数据流和创新点。",
		"result":     "展示主要实验、指标、消融和与 baseline 的差异。",
		"discussion": fmt.Sprintf("结合论文摘要或用户材料讨论局限：%s", emptyFallback(paperSummary, "数据规模、泛化性、复现成本和工程落地风险。")),
		"next_plan":  fmt.Sprintf("按 %s 汇报节奏给出 2-3 个下一步实验或工程验证。", duration),
	}
}

func jobMatch(jd, resume, preference string) map[string]any {
	score := 60 + int(math.Min(float64(len(commonWords(jd, resume))*8), 30))
	if preference != "" && strings.Contains(strings.ToLower(jd), strings.ToLower(preference)) {
		score += 5
	}
	if score > 95 {
		score = 95
	}
	return map[string]any{
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
	}
}

func interviewQuestions(project, jd, difficulty string) []string {
	if difficulty == "" {
		difficulty = "medium"
	}
	return []string{
		fmt.Sprintf("[%s] 你为什么把 %s 设计成 Agentic RAG，而不是普通 RAG？", difficulty, project),
		"Skill、Tool、MCP、Memory、RAG 的边界分别是什么？",
		"如果 MCP 工具调用失败，系统如何降级并保持 SSE 响应稳定？",
		"PaperChunker 为什么要保留 section metadata？后续如何做 parent-child chunking？",
		fmt.Sprintf("结合 JD：%s，你会如何把项目亮点改写成岗位相关表达？", emptyFallback(jd, "未提供 JD")),
	}
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

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func slug(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "，", "-", "。", "-")
	text = replacer.Replace(text)
	if text == "" {
		return "query"
	}
	return text
}
