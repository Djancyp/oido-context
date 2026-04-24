package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPHandler struct {
	db *SQLiteClient
}

type GetSessionContextArgs struct {
	SessionID string `json:"session_id" jsonschema:"Session ID to retrieve context for"`
	Limit     int    `json:"limit"      jsonschema:"Maximum events to return (default: 50)"`
}

type SearchContextArgs struct {
	Query string `json:"query" jsonschema:"Search terms to find in stored context"`
	Limit int    `json:"limit" jsonschema:"Maximum results to return (default: 20)"`
}

type ListSessionsArgs struct {
	Limit int `json:"limit" jsonschema:"Number of sessions to list (default: 10)"`
}

type GetRecentContextArgs struct {
	NSessions int `json:"n_sessions" jsonschema:"Number of most recent sessions to include (default: 3)"`
	Limit     int `json:"limit"      jsonschema:"Maximum events per session (default: 20)"`
}

func RunMCPServer() {
	db, err := NewSQLiteClient()
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	handler := &MCPHandler{db: db}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "oido-context",
		Version: "2.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_context",
		Description: "Retrieve stored conversation events for a specific session ID. Returns user prompts and assistant responses in chronological order.",
	}, handler.HandleGetSessionContext)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_context",
		Description: "Full-text search across all stored conversation history. Finds relevant context from any past session.",
	}, handler.HandleSearchContext)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List recent conversation sessions with their IDs, creation time, and last activity.",
	}, handler.HandleListSessions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_recent_context",
		Description: "Get conversation context from the N most recent sessions. Use this to expand the current session with recent history.",
	}, handler.HandleGetRecentContext)

	ctx := context.Background()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func (h *MCPHandler) HandleGetSessionContext(ctx context.Context, _ *mcp.CallToolRequest, args GetSessionContextArgs) (*mcp.CallToolResult, any, error) {
	if args.SessionID == "" {
		return errResult("session_id is required"), nil, nil
	}
	events, err := h.db.GetSessionEvents(ctx, args.SessionID, args.Limit)
	if err != nil {
		return errResult(fmt.Sprintf("query error: %v", err)), nil, nil
	}
	return textResult(formatEvents(events)), nil, nil
}

func (h *MCPHandler) HandleSearchContext(ctx context.Context, _ *mcp.CallToolRequest, args SearchContextArgs) (*mcp.CallToolResult, any, error) {
	if args.Query == "" {
		return errResult("query is required"), nil, nil
	}
	events, err := h.db.SearchEvents(ctx, args.Query, args.Limit)
	if err != nil {
		return errResult(fmt.Sprintf("search error: %v", err)), nil, nil
	}
	return textResult(formatEvents(events)), nil, nil
}

func (h *MCPHandler) HandleListSessions(ctx context.Context, _ *mcp.CallToolRequest, args ListSessionsArgs) (*mcp.CallToolResult, any, error) {
	sessions, err := h.db.ListSessions(ctx, args.Limit)
	if err != nil {
		return errResult(fmt.Sprintf("query error: %v", err)), nil, nil
	}
	if len(sessions) == 0 {
		return textResult("No sessions found."), nil, nil
	}
	var out string
	for _, s := range sessions {
		out += fmt.Sprintf("id=%s created=%s last_active=%s\n",
			s.ID,
			s.CreatedAt.Format("2006-01-02 15:04:05"),
			s.LastActive.Format("2006-01-02 15:04:05"),
		)
	}
	return textResult(out), nil, nil
}

func (h *MCPHandler) HandleGetRecentContext(ctx context.Context, _ *mcp.CallToolRequest, args GetRecentContextArgs) (*mcp.CallToolResult, any, error) {
	events, err := h.db.GetRecentSessionsEvents(ctx, args.NSessions, args.Limit)
	if err != nil {
		return errResult(fmt.Sprintf("query error: %v", err)), nil, nil
	}
	return textResult(formatEvents(events)), nil, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + msg}},
		IsError: true,
	}
}
