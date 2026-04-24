# Oido Context Extension

Automatically captures every user prompt and assistant response into a local SQLite database, keyed by session ID. Use the MCP tools below to retrieve or search past conversation history — useful for expanding the current session with prior context.

## Available Tools

### `get_session_context`
Retrieve all stored events for a specific session in chronological order.

**Parameters:**
- `session_id` (string, required): Session ID to look up
- `limit` (int, optional): Max events to return (default: 50)

**Returns:** User prompts and assistant responses for that session.

### `search_context`
Search stored conversation history by keyword.

**Parameters:**
- `query` (string, required): Search terms
- `limit` (int, optional): Max results (default: 20)

**Returns:** Matching events from any session, most recent first.

### `list_sessions`
List recent conversation sessions.

**Parameters:**
- `limit` (int, optional): Number of sessions (default: 10)

**Returns:** Session IDs with creation and last-active timestamps.

### `get_recent_context`
Load conversation context from the N most recent sessions.

**Parameters:**
- `n_sessions` (int, optional): Sessions to include (default: 3)
- `limit` (int, optional): Max events per session (default: 20)

**Returns:** Events from the most recent sessions, grouped chronologically.

## When to Use

- User says "remember what we were doing", "continue from last time", "expand context"
- User asks "what did we discuss about X" or "find where we talked about Y"
- User references a past session or prior work
- Starting a session that seems to continue previous work

## How Context Is Captured

Capture is automatic — no manual action needed:
- **UserPromptSubmit hook** → stores the user's prompt
- **Stop hook** → stores the assistant's final response

Both are keyed by session ID and stored in SQLite at:
`~/.config/oido/extensions/oido-context/context.db`

Override the path with the `OIDO_CONTEXT_DB` environment variable.
