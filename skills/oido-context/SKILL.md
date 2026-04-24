---
name: oido-context
description: Query stored session context — past prompts and assistant responses captured across Claude Code sessions
---

# Oido Context

Captures every user prompt (via UserPromptSubmit hook) and the last assistant response (via Stop hook) into a local SQLite database. Use the MCP tools below to retrieve or search that history.

## Available Tools

### `get_session_context`
Retrieve all events for a specific session in chronological order.

**Parameters:**
- `session_id` (string, required): The session ID to look up
- `limit` (int, optional): Max events to return (default: 50)

**Use when:** User references a specific past session, or you have a session ID.

### `search_context`
Full-text search across all stored conversation history.

**Parameters:**
- `query` (string, required): Search terms
- `limit` (int, optional): Max results (default: 20)

**Use when:** User asks "what did we discuss about X", "find where we talked about Y", or wants to locate past context by topic.

### `list_sessions`
List recent sessions with IDs and timestamps.

**Parameters:**
- `limit` (int, optional): Number of sessions (default: 10)

**Use when:** User asks "what sessions do I have", "show my recent conversations", or you need to browse available history.

### `get_recent_context`
Load context from the N most recent sessions — the main tool for expanding the current session.

**Parameters:**
- `n_sessions` (int, optional): Sessions to include (default: 3)
- `limit` (int, optional): Events per session (default: 20)

**Use when:** User says "remember what we were doing", "continue from last time", "expand context", or starts a session that seems to continue prior work.

## Trigger Phrases
- "what did we discuss", "previous session", "last time we talked"
- "remember when", "continue from", "expand context"
- "past conversations", "search history", "find where we"
- "what have we worked on"

## Storage
- DB: `~/.config/oido/extensions/oido-context/context.db` (SQLite)
- Override: set `OIDO_CONTEXT_DB` environment variable
- Capture: automatic via hooks — no manual action needed

## Related Commands
- `/get-session-context` — retrieve a specific session
- `/search-context` — search across all history
- `/recent-context` — load N most recent sessions
