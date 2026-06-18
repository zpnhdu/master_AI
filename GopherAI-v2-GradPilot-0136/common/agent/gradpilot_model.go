package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"GopherAI/common/memory"
	"GopherAI/common/skills"
	"GopherAI/common/tools"
	"GopherAI/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type GradPilotAgentModel struct {
	llm      model.ToolCallingChatModel
	username string
	loader   *skills.Loader
	store    *memory.Store
	registry *tools.Registry
	cfg      config.AgentConfig
}

type decision struct {
	Action string         `json:"action"`
	Answer string         `json:"answer"`
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args"`
}

func NewGradPilotAgentModel(ctx context.Context, username string) (*GradPilotAgentModel, error) {
	cfg := config.GetConfig()
	key := os.Getenv("OPENAI_API_KEY")

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.RagModelConfig.RagBaseUrl,
		Model:   cfg.RagModelConfig.RagChatModelName,
		APIKey:  key,
	})
	if err != nil {
		return nil, fmt.Errorf("create gradpilot model failed: %v", err)
	}

	loader := skills.NewLoader(cfg.AgentConfig.SkillDir)
	store := memory.NewStore(cfg.AgentConfig.MemoryDir)

	return &GradPilotAgentModel{
		llm:      llm,
		username: username,
		loader:   loader,
		store:    store,
		registry: tools.NewDefaultRegistry(username, loader, store),
		cfg:      cfg.AgentConfig,
	}, nil
}

func (m *GradPilotAgentModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	query := messages[len(messages)-1].Content
	bundle := m.buildContextBundle(query)
	firstResp, err := m.llm.Generate(ctx, replaceLastMessage(messages, m.buildDecisionPrompt(query, bundle)))
	if err != nil {
		return nil, fmt.Errorf("gradpilot first generate failed: %v", err)
	}

	decision, err := parseDecision(firstResp.Content)
	if err != nil {
		return m.llm.Generate(ctx, replaceLastMessage(messages, m.buildDirectPrompt(query, bundle)))
	}
	if decision.Action == "final" {
		return schema.AssistantMessage(decision.Answer, nil), nil
	}
	if decision.Action != "tool" {
		return schema.AssistantMessage(firstResp.Content, nil), nil
	}

	toolResult := m.callToolAsResult(ctx, decision)
	finalPrompt := m.buildFinalPrompt(query, bundle, decision, toolResult)
	finalResp, err := m.llm.Generate(ctx, replaceLastMessage(messages, finalPrompt))
	if err != nil {
		return nil, fmt.Errorf("gradpilot final generate failed: %v", err)
	}
	return finalResp, nil
}

func (m *GradPilotAgentModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb func(string)) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	query := messages[len(messages)-1].Content
	bundle := m.buildContextBundle(query)
	firstResp, err := m.llm.Generate(ctx, replaceLastMessage(messages, m.buildDecisionPrompt(query, bundle)))
	if err != nil {
		return "", fmt.Errorf("gradpilot first generate failed: %v", err)
	}

	decision, err := parseDecision(firstResp.Content)
	if err != nil {
		return m.streamPrompt(ctx, replaceLastMessage(messages, m.buildDirectPrompt(query, bundle)), cb)
	}
	if decision.Action == "final" {
		return decision.Answer, nil
	}
	if decision.Action != "tool" {
		return firstResp.Content, nil
	}

	toolResult := m.callToolAsResult(ctx, decision)
	finalPrompt := m.buildFinalPrompt(query, bundle, decision, toolResult)
	return m.streamPrompt(ctx, replaceLastMessage(messages, finalPrompt), cb)
}

func (m *GradPilotAgentModel) GetModelType() string {
	return "5"
}

func (m *GradPilotAgentModel) buildContextBundle(query string) string {
	intent := detectIntent(query)
	loadedSkills := m.loadIntentSkills(intent)
	skillPrompt := skills.BuildSkillPrompt(loadedSkills)

	var memoryText string
	if m.cfg.EnableMemory {
		categories := []string{"preferences"}
		if intent == "research" {
			categories = append(categories, "research")
		}
		if intent == "career" {
			categories = append(categories, "career")
		}
		for _, category := range categories {
			content, err := m.store.ReadMemory(m.username, category)
			if err == nil && strings.TrimSpace(content) != "" {
				memoryText += fmt.Sprintf("\n[%s]\n%s\n", category, content)
			}
		}
	}
	if strings.TrimSpace(memoryText) == "" {
		memoryText = "暂无长期偏好记忆。"
	}

	return fmt.Sprintf(`Intent: %s

Skill 方法论:
%s

Memory 长期偏好:
%s

可用内部工具:
%s`, intent, skillPrompt, memoryText, mustJSON(m.registry.List()))
}

func (m *GradPilotAgentModel) buildDecisionPrompt(query, bundle string) string {
	return fmt.Sprintf(`你是 GradPilot，一个面向研究生科研与求职场景的 Agentic RAG 助手。

职责边界:
- Skill 决定怎么想，Tool 决定能做什么，MCP 决定怎么连接外部能力。
- Skill 只作为方法论注入 prompt，不直接调用 MCP。
- MCP 只是工具执行层，是否调用工具由你决定。
- 当前只允许一次工具调用。

上下文:
%s

用户问题:
%s

请先判断是否需要工具。必须只返回严格 JSON，不要使用 markdown。

不需要工具:
{"action":"final","answer":"你的自然语言回答"}

需要工具:
{"action":"tool","tool":"mcp_call_tool","args":{"tool":"search_papers","args":{"query":"关键词","limit":5}}}

也可以调用 rag_search、memory_search、memory_write、paper_chunk、skill_list、skill_load、mcp_list_tools。`, bundle, query)
}

func (m *GradPilotAgentModel) buildDirectPrompt(query, bundle string) string {
	return fmt.Sprintf(`你是 GradPilot，一个科研与求职 Agent 助手。

上下文:
%s

用户问题:
%s

请直接给出清晰、可执行、不过度夸大的回答。`, bundle, query)
}

func (m *GradPilotAgentModel) buildFinalPrompt(query, bundle string, d decision, toolResult string) string {
	return fmt.Sprintf(`你是 GradPilot，一个科研与求职 Agent 助手。

上下文:
%s

用户问题:
%s

工具调用:
tool=%s
args=%s

工具结果:
%s

请基于工具结果进行二次总结。说明哪些是真实当前实现，哪些是 mock 或扩展点，不要夸大。`, bundle, query, d.Tool, mustJSON(d.Args), toolResult)
}

func (m *GradPilotAgentModel) callToolAsResult(ctx context.Context, d decision) string {
	if d.Args == nil {
		d.Args = map[string]any{}
	}
	result, err := m.registry.Call(ctx, d.Tool, d.Args)
	if err != nil {
		return "工具调用失败: " + err.Error()
	}
	return result
}

func (m *GradPilotAgentModel) loadIntentSkills(intent string) []skills.Skill {
	if !m.cfg.EnableSkill {
		return nil
	}
	names := []string{}
	switch intent {
	case "research":
		names = []string{"paper-reading", "literature-review", "code-reproduction", "research-presentation"}
	case "career":
		names = []string{"resume-optimization", "interview-project"}
	default:
		names = []string{"paper-reading", "resume-optimization"}
	}

	out := make([]skills.Skill, 0, len(names))
	for _, name := range names {
		skill, err := m.loader.LoadSkill(name)
		if err == nil {
			out = append(out, *skill)
		}
	}
	return out
}

func (m *GradPilotAgentModel) streamPrompt(ctx context.Context, messages []*schema.Message, cb func(string)) (string, error) {
	stream, err := m.llm.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("gradpilot stream failed: %v", err)
	}
	defer stream.Close()

	var full strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("gradpilot stream recv failed: %v", err)
		}
		if msg.Content != "" {
			full.WriteString(msg.Content)
			cb(msg.Content)
		}
	}
	return full.String(), nil
}

func replaceLastMessage(messages []*schema.Message, content string) []*schema.Message {
	out := make([]*schema.Message, len(messages))
	copy(out, messages)
	out[len(out)-1] = &schema.Message{
		Role:    schema.User,
		Content: content,
	}
	return out
}

func parseDecision(raw string) (decision, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var d decision
	if err := json.Unmarshal([]byte(raw), &d); err == nil {
		return d, nil
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &d); err == nil {
			return d, nil
		}
	}
	return decision{}, fmt.Errorf("invalid decision JSON")
}

func detectIntent(query string) string {
	lower := strings.ToLower(query)
	researchWords := []string{"论文", "paper", "arxiv", "github", "复现", "组会", "实验", "method", "research", "literature"}
	careerWords := []string{"简历", "岗位", "面试", "jd", "resume", "career", "offer", "求职", "招聘"}
	for _, word := range researchWords {
		if strings.Contains(lower, word) {
			return "research"
		}
	}
	for _, word := range careerWords {
		if strings.Contains(lower, word) {
			return "career"
		}
	}
	return "general"
}

func mustJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}
