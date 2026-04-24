package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestDB creates a fresh SQLiteClient backed by a temp file.
func newTestDB(t *testing.T) *SQLiteClient {
	t.Helper()
	t.Setenv("OIDO_CONTEXT_DB", filepath.Join(t.TempDir(), "test.db"))
	db, err := NewSQLiteClient()
	if err != nil {
		t.Fatalf("NewSQLiteClient: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// text extracts the first TextContent from a CallToolResult.
func text(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] is not *mcp.TextContent: %T", result.Content[0])
	}
	return tc.Text
}

// seed inserts a session + user prompt + assistant response.
func seed(t *testing.T, db *SQLiteClient, sessionID, prompt, response string) {
	t.Helper()
	ctx := context.Background()
	if err := db.UpsertSession(ctx, sessionID); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}
	if err := db.InsertEvent(ctx, Event{SessionID: sessionID, EventType: "UserPromptSubmit", Role: "user", Content: prompt, Metadata: "{}"}); err != nil {
		t.Fatalf("InsertEvent user: %v", err)
	}
	if err := db.InsertEvent(ctx, Event{SessionID: sessionID, EventType: "Stop", Role: "assistant", Content: response, Metadata: "{}"}); err != nil {
		t.Fatalf("InsertEvent assistant: %v", err)
	}
}

// newHandler builds an MCPHandler over a fresh test DB.
func newHandler(t *testing.T) *MCPHandler {
	t.Helper()
	return &MCPHandler{db: newTestDB(t)}
}

// ---- SQLite layer ----

func TestUpsertSession(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.UpsertSession(ctx, "s1"); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := db.UpsertSession(ctx, "s1"); err != nil {
		t.Fatalf("second upsert (update): %v", err)
	}

	sessions, err := db.ListSessions(ctx, 10)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "s1" {
		t.Fatalf("want 1 session 's1', got %+v", sessions)
	}
}

func TestInsertAndGetSessionEvents(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "s1", "hello world", "hi there")

	events, err := db.GetSessionEvents(context.Background(), "s1", 0)
	if err != nil {
		t.Fatalf("GetSessionEvents: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 events, got %d", len(events))
	}
	if events[0].Role != "user" || events[0].Content != "hello world" {
		t.Errorf("event[0] wrong: %+v", events[0])
	}
	if events[1].Role != "assistant" || events[1].Content != "hi there" {
		t.Errorf("event[1] wrong: %+v", events[1])
	}
}

func TestGetSessionEvents_Limit(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "s1", "q1", "a1")
	seed(t, db, "s1", "q2", "a2")

	events, err := db.GetSessionEvents(context.Background(), "s1", 1)
	if err != nil {
		t.Fatalf("GetSessionEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 (limit), got %d", len(events))
	}
}

func TestGetSessionEvents_UnknownSession(t *testing.T) {
	db := newTestDB(t)
	events, err := db.GetSessionEvents(context.Background(), "no-such", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("want 0 events, got %d", len(events))
	}
}

func TestSearchEvents(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "s1", "deploy kubernetes cluster", "done")
	seed(t, db, "s2", "fix login bug", "patched")

	events, err := db.SearchEvents(context.Background(), "kubernetes", 10)
	if err != nil {
		t.Fatalf("SearchEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 result, got %d", len(events))
	}
	if !strings.Contains(events[0].Content, "kubernetes") {
		t.Errorf("unexpected content: %s", events[0].Content)
	}
}

func TestSearchEvents_NoMatch(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "s1", "hello world", "hi")

	events, err := db.SearchEvents(context.Background(), "xyzzy", 10)
	if err != nil {
		t.Fatalf("SearchEvents: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("want 0 results, got %d", len(events))
	}
}

func TestListSessions_Order(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	for _, id := range []string{"s-a", "s-b", "s-c"} {
		seed(t, db, id, "p", "r")
	}
	// Re-touch s-a to make it the most recent.
	if err := db.UpsertSession(ctx, "s-a"); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	sessions, err := db.ListSessions(ctx, 10)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("want 3, got %d", len(sessions))
	}
	if sessions[0].ID != "s-a" {
		t.Errorf("want s-a first (most recent), got %s", sessions[0].ID)
	}
}

func TestListSessions_Limit(t *testing.T) {
	db := newTestDB(t)
	for _, id := range []string{"s1", "s2", "s3", "s4", "s5"} {
		seed(t, db, id, "p", "r")
	}
	sessions, err := db.ListSessions(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("want 3, got %d", len(sessions))
	}
}

func TestGetRecentSessionsEvents(t *testing.T) {
	db := newTestDB(t)
	seed(t, db, "old", "old prompt", "old response")
	seed(t, db, "new", "new prompt", "new response")

	events, err := db.GetRecentSessionsEvents(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetRecentSessionsEvents: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 events (1 session × 2 turns), got %d", len(events))
	}
	if events[0].SessionID != "new" {
		t.Errorf("want session 'new', got %s", events[0].SessionID)
	}
}

// ---- MCP handlers ----

func TestHandleGetSessionContext_MissingID(t *testing.T) {
	h := newHandler(t)
	result, _, err := h.HandleGetSessionContext(context.Background(), nil, GetSessionContextArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("want IsError=true for missing session_id")
	}
}

func TestHandleGetSessionContext_Found(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "s1", "what is go?", "Go is a language.")

	result, _, err := h.HandleGetSessionContext(context.Background(), nil, GetSessionContextArgs{SessionID: "s1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", text(t, result))
	}
	out := text(t, result)
	if !strings.Contains(out, "what is go?") {
		t.Errorf("want prompt in output, got: %s", out)
	}
	if !strings.Contains(out, "Go is a language.") {
		t.Errorf("want response in output, got: %s", out)
	}
}

func TestHandleGetSessionContext_Empty(t *testing.T) {
	h := newHandler(t)
	result, _, err := h.HandleGetSessionContext(context.Background(), nil, GetSessionContextArgs{SessionID: "no-such"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result")
	}
	if !strings.Contains(text(t, result), "No events found") {
		t.Errorf("want 'No events found', got: %s", text(t, result))
	}
}

func TestHandleSearchContext_MissingQuery(t *testing.T) {
	h := newHandler(t)
	result, _, err := h.HandleSearchContext(context.Background(), nil, SearchContextArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("want IsError=true for missing query")
	}
}

func TestHandleSearchContext_Found(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "s1", "deploy to production", "deployed successfully")
	seed(t, h.db, "s2", "fix the login page", "login fixed")

	result, _, err := h.HandleSearchContext(context.Background(), nil, SearchContextArgs{Query: "deploy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := text(t, result)
	if !strings.Contains(out, "deploy") {
		t.Errorf("want 'deploy' in results, got: %s", out)
	}
	if strings.Contains(out, "login") {
		t.Errorf("'login' should not appear in deploy search: %s", out)
	}
}

func TestHandleSearchContext_NoMatch(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "s1", "hello world", "hi")

	result, _, err := h.HandleSearchContext(context.Background(), nil, SearchContextArgs{Query: "xyzzy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text(t, result), "No events found") {
		t.Errorf("want 'No events found', got: %s", text(t, result))
	}
}

func TestHandleListSessions_Empty(t *testing.T) {
	h := newHandler(t)
	result, _, err := h.HandleListSessions(context.Background(), nil, ListSessionsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text(t, result), "No sessions found") {
		t.Errorf("want 'No sessions found', got: %s", text(t, result))
	}
}

func TestHandleListSessions_Found(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "sess-alpha", "p", "r")
	seed(t, h.db, "sess-beta", "p", "r")

	result, _, err := h.HandleListSessions(context.Background(), nil, ListSessionsArgs{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := text(t, result)
	if !strings.Contains(out, "sess-alpha") {
		t.Errorf("want sess-alpha in output, got: %s", out)
	}
	if !strings.Contains(out, "sess-beta") {
		t.Errorf("want sess-beta in output, got: %s", out)
	}
}

func TestHandleGetRecentContext_Empty(t *testing.T) {
	h := newHandler(t)
	result, _, err := h.HandleGetRecentContext(context.Background(), nil, GetRecentContextArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text(t, result), "No events found") {
		t.Errorf("want 'No events found', got: %s", text(t, result))
	}
}

func TestHandleGetRecentContext_Found(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "s1", "first prompt", "first response")
	seed(t, h.db, "s2", "second prompt", "second response")

	result, _, err := h.HandleGetRecentContext(context.Background(), nil, GetRecentContextArgs{NSessions: 2, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := text(t, result)
	if !strings.Contains(out, "second prompt") {
		t.Errorf("want second session in output, got: %s", out)
	}
}

func TestHandleGetRecentContext_NSessions(t *testing.T) {
	h := newHandler(t)
	seed(t, h.db, "s1", "old prompt", "old response")
	seed(t, h.db, "s2", "recent prompt", "recent response")

	result, _, err := h.HandleGetRecentContext(context.Background(), nil, GetRecentContextArgs{NSessions: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := text(t, result)
	if strings.Contains(out, "old prompt") {
		t.Errorf("old session should be excluded with NSessions=1, got: %s", out)
	}
	if !strings.Contains(out, "recent prompt") {
		t.Errorf("want recent session in output, got: %s", out)
	}
}

// ---- Store / hook layer ----

func TestStoreUserPromptSubmit(t *testing.T) {
	t.Setenv("OIDO_CONTEXT_DB", filepath.Join(t.TempDir(), "test.db"))

	payload := `{"session_id":"hook-sess","prompt":"hook test prompt"}`
	if err := storeFromReader(strings.NewReader(payload), "UserPromptSubmit"); err != nil {
		t.Fatalf("storeFromReader: %v", err)
	}

	db, _ := NewSQLiteClient()
	defer db.Close()
	events, _ := db.GetSessionEvents(context.Background(), "hook-sess", 0)
	if len(events) != 1 || events[0].Role != "user" || events[0].Content != "hook test prompt" {
		t.Fatalf("wrong events: %+v", events)
	}
}

func TestStoreStop(t *testing.T) {
	t.Setenv("OIDO_CONTEXT_DB", filepath.Join(t.TempDir(), "test.db"))

	payload := `{"session_id":"hook-sess","reason":"final_response","response":"the assistant reply"}`
	if err := storeFromReader(strings.NewReader(payload), "Stop"); err != nil {
		t.Fatalf("storeFromReader: %v", err)
	}

	db, _ := NewSQLiteClient()
	defer db.Close()
	events, _ := db.GetSessionEvents(context.Background(), "hook-sess", 0)
	if len(events) != 1 || events[0].Role != "assistant" || events[0].Content != "the assistant reply" {
		t.Fatalf("wrong events: %+v", events)
	}
}

func TestStoreSkipsEmptyPrompt(t *testing.T) {
	t.Setenv("OIDO_CONTEXT_DB", filepath.Join(t.TempDir(), "test.db"))

	payload := `{"session_id":"hook-sess","prompt":""}`
	if err := storeFromReader(strings.NewReader(payload), "UserPromptSubmit"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db, _ := NewSQLiteClient()
	defer db.Close()
	events, _ := db.GetSessionEvents(context.Background(), "hook-sess", 0)
	if len(events) != 0 {
		t.Fatalf("want 0 events for empty prompt, got %d", len(events))
	}
}

func TestStoreSkipsEmptyResponse(t *testing.T) {
	t.Setenv("OIDO_CONTEXT_DB", filepath.Join(t.TempDir(), "test.db"))

	payload := `{"session_id":"hook-sess","response":""}`
	if err := storeFromReader(strings.NewReader(payload), "Stop"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db, _ := NewSQLiteClient()
	defer db.Close()
	events, _ := db.GetSessionEvents(context.Background(), "hook-sess", 0)
	if len(events) != 0 {
		t.Fatalf("want 0 events for empty response, got %d", len(events))
	}
}

// ---- DB path ----

func TestDBPathEnvOverride(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "custom.db")
	t.Setenv("OIDO_CONTEXT_DB", custom)
	if got := dbPath(); got != custom {
		t.Errorf("want %s, got %s", custom, got)
	}
}

func TestDBPathDefault(t *testing.T) {
	t.Setenv("OIDO_CONTEXT_DB", "")
	want := filepath.Join(os.Getenv("HOME"), ".config", "oido", "extensions", "oido-context", "context.db")
	if got := dbPath(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
