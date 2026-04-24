# Oido PostgreSQL Plugin — PostgreSQL MCP Extension

This is a **PostgreSQL extension plugin** for Oido Studio. It provides tools for querying PostgreSQL databases, listing tables, and describing schemas via the MCP protocol.

## Features

- **`execute_sql`** — Execute SELECT queries against PostgreSQL with safety blocking of destructive operations
- **`list_tables`** — List all tables in the current database
- **`describe_table`** — Show column names, types, and constraints for a table

## Building

```bash
make build
# or
go build -o oido-postgres-mcp .
```

## Installing

Place the plugin in `plugins/oido-postgres/`. The plugin manager discovers it automatically on startup.

## Environment Variables

Set these before starting Oido Studio:

| Variable | Required | Description |
|----------|----------|-------------|
| `POSTGRES_HOST` | Yes | PostgreSQL server hostname/IP |
| `POSTGRES_PORT` | No | PostgreSQL port (default: 5432) |
| `POSTGRES_DATABASE` | Yes | Database name |
| `POSTGRES_USER` | Yes | Database username |
| `POSTGRES_PASSWORD` | No | Database password |

## SQL Function Permissions

Control which SQL operations are allowed via environment variables. By default, only SELECT is enabled for safety.

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_ALLOW_SELECT` | `true` | Allow SELECT queries |
| `POSTGRES_ALLOW_INSERT` | `false` | Allow INSERT statements |
| `POSTGRES_ALLOW_UPDATE` | `false` | Allow UPDATE statements |
| `POSTGRES_ALLOW_DELETE` | `false` | Allow DELETE statements |
| `POSTGRES_ALLOW_CREATE` | `false` | Allow CREATE statements (tables, indexes, etc.) |
| `POSTGRES_ALLOW_ALTER` | `false` | Allow ALTER statements (modify tables, etc.) |
| `POSTGRES_ALLOW_DROP` | `false` | Allow DROP statements (delete tables, etc.) |
| `POSTGRES_ALLOW_TRUNCATE` | `false` | Allow TRUNCATE statements |

### Examples

**Enable read-write access:**
```bash
export POSTGRES_ALLOW_SELECT=true
export POSTGRES_ALLOW_INSERT=true
export POSTGRES_ALLOW_UPDATE=true
export POSTGRES_ALLOW_DELETE=true
```

**Enable full database access (including DDL):**
```bash
export POSTGRES_ALLOW_SELECT=true
export POSTGRES_ALLOW_INSERT=true
export POSTGRES_ALLOW_UPDATE=true
export POSTGRES_ALLOW_DELETE=true
export POSTGRES_ALLOW_CREATE=true
export POSTGRES_ALLOW_ALTER=true
export POSTGRES_ALLOW_DROP=true
export POSTGRES_ALLOW_TRUNCATE=true
```

**Read-only mode (default):**
```bash
# Only SELECT enabled (default behavior)
export POSTGRES_ALLOW_SELECT=true
```

## Testing

Start Oido Studio, then in chat:

```
List all tables in the database
```

Or run commands:

```bash
/list-tables
/describe-table table=users
/sql-query query="SELECT * FROM public.users LIMIT 10"
```

## Architecture

```
oido-postgres/
├── plugin.json              # Plugin manifest
├── qwen-extension.json      # Qwen CLI extension config
├── oido-postgres-mcp        # Compiled binary
├── main.go                  # Entry point
├── mcp_server.go            # MCP tool handlers
├── postgres.go              # PostgreSQL client & query logic
├── Makefile                 # Build helper
├── QWEN.md                  # LLM context file
├── commands/                # Custom CLI commands
│   ├── sql-query.toml
│   ├── list-tables.toml
│   └── describe-table.toml
├── skills/
│   └── oido-postgres/
│       └── SKILL.md         # Skill documentation
└── README.md                # This file
```

## Safety

- **Configurable SQL permissions**: Each SQL operation type (SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, TRUNCATE) can be individually enabled/disabled via environment variables
- **Default read-only**: Only SELECT queries enabled by default
- **Row limits**: Default 100 rows for SELECT queries
- **Connection pooling**: 10 max open, 5 idle connections

## Creating Your Own Database Plugin

Copy this directory and modify:

1. `plugin.json` — Update `id`, `name`, `description`, `binary`, `capabilities`, `config_schema`
2. `go.mod` — Update module path and database driver
3. `postgres.go` — Replace with your database client logic
4. `mcp_server.go` — Update tool names and handlers
5. Build: `make build`
