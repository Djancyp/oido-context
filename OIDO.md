# Oido PostgreSQL Extension

Execute SQL queries, list tables, and describe PostgreSQL database schemas. This extension provides MCP tools for database exploration and query execution.

## Available Tools

### `execute_sql`
Execute a SQL query against the PostgreSQL database.

**Parameters:**
- `query` (string, required): SQL query to execute (SELECT only for safety)
- `limit` (number, optional): Maximum rows to return (default: 100)

**Returns:** Formatted table results with column headers and row counts.

### `list_tables`
List all tables in the PostgreSQL database.

**Parameters:** None

**Returns:** List of schema.table names.

### `describe_table`
Describe a table's columns, types, and constraints.

**Parameters:**
- `schema` (string, optional): Database schema name (default: public)
- `table` (string, required): Table name to describe

**Returns:** Column names, data types, nullable status, and defaults.

## Example Usage

```
User: What tables are in the database?

Assistant: I'll list all tables in your PostgreSQL database.

[Uses list_tables tool]

Tables (5):

public.users
public.orders
public.products
...
```

```
User: Show me the first 10 users

Assistant: I'll query the users table for you.

[Uses execute_sql tool with query: "SELECT * FROM public.users", limit: 10]

Query returned 5 columns

id | name | email | created_at | status
----------------------------------------
1 | Alice | alice@example.com | 2024-01-01 | active
2 | Bob | bob@example.com | 2024-01-02 | active

Total rows: 10
```

## When to Use

- User asks to query database or run SQL
- User wants to explore database schema
- User asks about tables, columns, or data
- User wants to analyze database structure
- User asks for data analysis or reporting

## Notes

- **Configurable SQL permissions**: Each SQL operation type (SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, TRUNCATE) can be individually enabled/disabled via environment variables
- **Default read-only**: Only SELECT queries enabled by default
- **Environment variables for permissions**:
  - `POSTGRES_ALLOW_SELECT` (default: true)
  - `POSTGRES_ALLOW_INSERT` (default: false)
  - `POSTGRES_ALLOW_UPDATE` (default: false)
  - `POSTGRES_ALLOW_DELETE` (default: false)
  - `POSTGRES_ALLOW_CREATE` (default: false)
  - `POSTGRES_ALLOW_ALTER` (default: false)
  - `POSTGRES_ALLOW_DROP` (default: false)
  - `POSTGRES_ALLOW_TRUNCATE` (default: false)
- **Row limits**: Default 100 rows for SELECT queries
- **Connection pooling**: 10 max open, 5 idle, 5 min lifetime
- Uses environment variables for connection: POSTGRES_HOST, POSTGRES_PORT, POSTGRES_DATABASE, POSTGRES_USER, POSTGRES_PASSWORD
