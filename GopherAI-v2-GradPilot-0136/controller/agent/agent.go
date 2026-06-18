package agent

import (
	"net/http"

	"GopherAI/common/memory"
	"GopherAI/common/skills"
	"GopherAI/common/tools"
	"GopherAI/config"

	"github.com/gin-gonic/gin"
)

type callToolRequest struct {
	Name string         `json:"name" binding:"required"`
	Args map[string]any `json:"args"`
}

type writeMemoryRequest struct {
	Category string `json:"category" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

type searchMemoryRequest struct {
	Query string `json:"query" binding:"required"`
}

func Health(c *gin.Context) {
	loader, _, registry := deps(c.GetString("userName"))
	c.JSON(http.StatusOK, gin.H{
		"agent":    true,
		"mode":     "GradPilot",
		"skills":   len(loader.ListSkills()),
		"tools":    len(registry.List()),
		"mcpTools": len(tools.MCPToolInfos()),
	})
}

func ListSkills(c *gin.Context) {
	loader, _, _ := deps(c.GetString("userName"))
	c.JSON(http.StatusOK, gin.H{"skills": loader.ListSkills()})
}

func ListTools(c *gin.Context) {
	_, _, registry := deps(c.GetString("userName"))
	c.JSON(http.StatusOK, gin.H{"tools": registry.List()})
}

func CallTool(c *gin.Context) {
	req := new(callToolRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _, registry := deps(c.GetString("userName"))
	result, err := registry.Call(c.Request.Context(), req.Name, req.Args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func ReadMemory(c *gin.Context) {
	_, store, _ := deps(c.GetString("userName"))
	content, err := store.ReadMemory(userName(c), c.Param("category"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

func WriteMemory(c *gin.Context) {
	req := new(writeMemoryRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, store, _ := deps(c.GetString("userName"))
	if err := store.WriteMemory(userName(c), req.Category, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SearchMemory(c *gin.Context) {
	req := new(searchMemoryRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, store, _ := deps(c.GetString("userName"))
	results, err := store.SearchMemory(userName(c), req.Query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

func deps(username string) (*skills.Loader, *memory.Store, *tools.Registry) {
	cfg := config.GetConfig().AgentConfig
	loader := skills.NewLoader(cfg.SkillDir)
	store := memory.NewStore(cfg.MemoryDir)
	registry := tools.NewDefaultRegistry(username, loader, store)
	return loader, store, registry
}

func userName(c *gin.Context) string {
	name := c.GetString("userName")
	if name == "" {
		return "demo"
	}
	return name
}
