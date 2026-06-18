package mcp

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"gradpilot-tool-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	RegisterWeatherTools(mcpServer)
	RegisterPaperTools(mcpServer)
	RegisterGitHubTools(mcpServer)
	RegisterJobTools(mcpServer)
	RegisterWebTools(mcpServer)
	RegisterResearchTools(mcpServer)
	RegisterCareerTools(mcpServer)

	return mcpServer
}

func StartServer(httpAddr string) error {
	mcpServer := NewMCPServer()
	httpServer := server.NewStreamableHTTPServer(mcpServer)
	log.Printf("HTTP MCP server listening on %s/mcp", httpAddr)
	return httpServer.Start(httpAddr)
}
