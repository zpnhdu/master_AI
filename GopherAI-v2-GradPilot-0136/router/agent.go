package router

import (
	"GopherAI/controller/agent"

	"github.com/gin-gonic/gin"
)

func AgentRouter(r *gin.RouterGroup) {
	r.GET("/health", agent.Health)
	r.GET("/skills", agent.ListSkills)
	r.GET("/tools", agent.ListTools)
	r.POST("/tools/call", agent.CallTool)
	r.GET("/memory/:category", agent.ReadMemory)
	r.POST("/memory", agent.WriteMemory)
	r.POST("/memory/search", agent.SearchMemory)
}
