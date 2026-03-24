package mcp

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/huypham/agent-tutor/internal/config"
	"github.com/huypham/agent-tutor/internal/store"
)

// EmptyInput is used for tools that take no input parameters.
type EmptyInput struct{}

// IntensityInput is the input for the set_coaching_intensity tool.
type IntensityInput struct {
	Intensity string `json:"intensity" jsonschema:"The coaching intensity level: proactive, on-demand, or silent"`
}

// Server wraps the MCP server with agent-tutor tools.
type Server struct {
	server   *gomcp.Server
	handlers *ToolHandlers
}

// NewServer creates a new MCP server with all tutor tools registered.
func NewServer(s *store.Store, cfg *config.Config) *Server {
	handlers := NewToolHandlers(s, cfg)

	server := gomcp.NewServer(
		&gomcp.Implementation{
			Name:    "agent-tutor",
			Version: "0.1.0",
		},
		&gomcp.ServerOptions{
			Instructions: BuildInstructions(cfg),
		},
	)

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_student_context",
		Description: "Get a summary of recent student activity including file changes, terminal output, and git operations",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.GetStudentContext()
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_recent_file_changes",
		Description: "Get recent file changes with diffs",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.GetRecentFileChanges()
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_terminal_activity",
		Description: "Get recent terminal activity from the student's pane",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.GetTerminalActivity()
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_git_activity",
		Description: "Get recent git activity including commits and status changes",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.GetGitActivity()
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_coaching_config",
		Description: "Get the current coaching configuration (intensity and level)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.GetCoachingConfig()
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "set_coaching_intensity",
		Description: "Set the coaching intensity level",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input IntensityInput) (*gomcp.CallToolResult, any, error) {
		result := handlers.SetCoachingIntensity(input.Intensity)
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: result}},
		}, nil, nil
	})

	return &Server{server: server, handlers: handlers}
}

// Run starts the MCP server on stdio transport.
func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &gomcp.StdioTransport{})
}
