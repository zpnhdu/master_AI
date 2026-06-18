package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"GopherAI/common/memory"
	"GopherAI/common/rag"
	"GopherAI/common/skills"
)

type Tool struct {
	Name        string
	Description string
	Parameters  any
	Call        func(ctx context.Context, args map[string]any) (string, error)
}

type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func NewDefaultRegistry(userID string, loader *skills.Loader, store *memory.Store) *Registry {
	registry := NewRegistry()
	registry.Register(defaultRAGSearchTool(userID))
	registry.Register(defaultPaperChunkTool())
	registry.Register(defaultMemorySearchTool(userID, store))
	registry.Register(defaultMemoryWriteTool(userID, store))
	registry.Register(defaultSkillListTool(loader))
	registry.Register(defaultSkillLoadTool(loader))
	registry.Register(defaultMCPListToolsTool())
	registry.Register(defaultMCPCallTool())
	return registry
}

func (r *Registry) Register(tool Tool) error {
	if strings.TrimSpace(tool.Name) == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Call == nil {
		return fmt.Errorf("tool %s call handler is required", tool.Name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
	return nil
}

func (r *Registry) List() []ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		out = append(out, ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}
	return out
}

func (r *Registry) Call(ctx context.Context, name string, args map[string]any) (string, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("tool %s not found", name)
	}
	if args == nil {
		args = map[string]any{}
	}
	return tool.Call(ctx, args)
}

func defaultRAGSearchTool(userID string) Tool {
	return Tool{
		Name:        "rag_search",
		Description: "查询用户上传文档对应的本地 RAG；不可用时返回 mock fallback。",
		Parameters:  map[string]any{"query": "string required"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			query, err := getStringArg(args, "query", true)
			if err != nil {
				return "", err
			}
			ragQuery, err := rag.NewRAGQuery(ctx, userID)
			if err != nil {
				return fmt.Sprintf("RAG fallback: 当前用户暂无可检索文档或向量库不可用。query=%q，后续可接入已上传论文/简历文档。", query), nil
			}
			docs, err := ragQuery.RetrieveDocuments(ctx, query)
			if err != nil {
				return fmt.Sprintf("RAG fallback: 检索失败：%v。query=%q", err, query), nil
			}
			var b strings.Builder
			for i, doc := range docs {
				b.WriteString(fmt.Sprintf("[文档 %d]\n%s\n\n", i+1, doc.Content))
			}
			return b.String(), nil
		},
	}
}

func defaultPaperChunkTool() Tool {
	return Tool{
		Name:        "paper_chunk",
		Description: "按论文常见章节切块，长章节使用 sliding window。",
		Parameters:  map[string]any{"text": "string required"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			text, err := getStringArg(args, "text", true)
			if err != nil {
				return "", err
			}
			chunks := rag.ChunkBySection(text)
			return marshalJSON(chunks)
		},
	}
}

func defaultMemorySearchTool(userID string, store *memory.Store) Tool {
	return Tool{
		Name:        "memory_search",
		Description: "查询用户长期偏好记忆。",
		Parameters:  map[string]any{"query": "string required"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			query, err := getStringArg(args, "query", true)
			if err != nil {
				return "", err
			}
			results, err := store.SearchMemory(userID, query)
			if err != nil {
				return "", err
			}
			return marshalJSON(results)
		},
	}
}

func defaultMemoryWriteTool(userID string, store *memory.Store) Tool {
	return Tool{
		Name:        "memory_write",
		Description: "写入用户长期偏好记忆。",
		Parameters:  map[string]any{"category": "profile/research/career/preferences", "content": "string required"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			category, err := getStringArg(args, "category", true)
			if err != nil {
				return "", err
			}
			content, err := getStringArg(args, "content", true)
			if err != nil {
				return "", err
			}
			if err := store.WriteMemory(userID, category, content); err != nil {
				return "", err
			}
			return "memory written", nil
		},
	}
}

func defaultSkillListTool(loader *skills.Loader) Tool {
	return Tool{
		Name:        "skill_list",
		Description: "列出可用 Skills。",
		Parameters:  map[string]any{},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			return marshalJSON(loader.ListSkills())
		},
	}
}

func defaultSkillLoadTool(loader *skills.Loader) Tool {
	return Tool{
		Name:        "skill_load",
		Description: "加载指定 Skill 内容。",
		Parameters:  map[string]any{"name": "string required"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			name, err := getStringArg(args, "name", true)
			if err != nil {
				return "", err
			}
			skill, err := loader.LoadSkill(name)
			if err != nil {
				return "", err
			}
			return skill.RawContent, nil
		},
	}
}

func defaultMCPListToolsTool() Tool {
	return Tool{
		Name:        "mcp_list_tools",
		Description: "列出 GradPilot 内置 MCP Server 工具。",
		Parameters:  map[string]any{},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			return marshalJSON(MCPToolInfos())
		},
	}
}

func defaultMCPCallTool() Tool {
	return Tool{
		Name:        "mcp_call_tool",
		Description: "调用 GradPilot 内置 MCP Server 工具。",
		Parameters:  map[string]any{"tool": "string required", "args": "object optional"},
		Call: func(ctx context.Context, args map[string]any) (string, error) {
			toolName, err := getStringArg(args, "tool", true)
			if err != nil {
				return "", err
			}
			toolArgs := map[string]any{}
			if raw, ok := args["args"]; ok {
				if converted, ok := raw.(map[string]any); ok {
					toolArgs = converted
				} else {
					return "", fmt.Errorf("args must be an object")
				}
			}
			return CallBuiltinMCPTool(ctx, toolName, toolArgs)
		},
	}
}

func getStringArg(args map[string]any, key string, required bool) (string, error) {
	value, ok := args[key]
	if !ok {
		if required {
			return "", fmt.Errorf("%s is required", key)
		}
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be string", key)
	}
	text = strings.TrimSpace(text)
	if required && text == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return text, nil
}

func marshalJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
