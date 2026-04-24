# oido-context

Session context capture and retrieval for Oido Studio. Automatically stores user prompts and assistant responses per session in a local SQLite database, then exposes MCP tools so Claude can query that history to expand its context across sessions.

## How It Works

```
UserPromptSubmit hook → oido-context-mcp store UserPromptSubmit → SQLite
Stop hook             → oido-context-mcp store Stop             → SQLite
MCP server            → get_session_context / search_context / list_sessions / get_recent_context
```

The binary runs in two modes:
- **MCP server** (no args) — started by Oido Studio, serves query tools via stdio
- **Store mode** (`store <EventType>`) — called by hooks, reads JSON from stdin and writes to SQLite

## MCP Tools

| Tool | Args | Description |
|------|------|-------------|
| `get_session_context` | `session_id`, `limit` | All events for a session, chronological |
| `search_context` | `query`, `limit` | LIKE search across all stored content |
| `list_sessions` | `limit` | Recent sessions with timestamps |
| `get_recent_context` | `n_sessions`, `limit` | Events from N most recent sessions |

## Commands

| Command | Description |
|---------|-------------|
| `/get-session-context` | Retrieve a specific session |
| `/search-context` | Search across all history |
| `/recent-context` | Load N most recent sessions |

## Building

```bash
make build
```

Produces `oido-context-mcp` (static binary, no CGO).

## Database

SQLite database is stored at:

```
~/.config/oido/extensions/oido-context/context.db
```

Override with `OIDO_CONTEXT_DB` environment variable.

## Settings

| Setting | Description |
|---------|-------------|
| `OIDO_CONTEXT_DB` | Custom path for the SQLite DB file |

## Running Tests

```bash
go test ./...
```

## Project Structure

```
oido-context/
├── main.go               # CLI dispatch (store mode vs MCP server)
├── mcp_server.go         # MCP tool handlers
├── sqlite.go             # SQLite client, schema, queries
├── store.go              # Hook payload parsing and storage
├── context_test.go       # Tests for all layers
├── Makefile
├── oido-extension.json   # Extension manifest (hooks + MCP server)
├── OIDO.md               # LLM context file
├── commands/
│   ├── get-session-context.toml
│   ├── search-context.toml
│   └── recent-context.toml
└── skills/
    └── oido-context/
        └── SKILL.md
```
