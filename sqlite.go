package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteClient struct {
	db *sql.DB
}

type Session struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

type Event struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	EventType string    `json:"event_type"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

func dbPath() string {
	if p := os.Getenv("OIDO_CONTEXT_DB"); p != "" {
		return p
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "oido", "extensions", "oido-context", "context.db")
}

var schema = []string{
	`CREATE TABLE IF NOT EXISTS sessions (
		id          TEXT PRIMARY KEY,
		created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
		last_active DATETIME NOT NULL DEFAULT (datetime('now'))
	)`,
	`CREATE TABLE IF NOT EXISTS events (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id  TEXT NOT NULL REFERENCES sessions(id),
		event_type  TEXT NOT NULL,
		role        TEXT NOT NULL,
		content     TEXT NOT NULL,
		metadata    TEXT NOT NULL DEFAULT '{}',
		created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
	)`,
	`CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_time    ON events(created_at DESC)`,
}

func NewSQLiteClient() (*SQLiteClient, error) {
	p := dbPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", p)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Single writer at a time is fine; WAL allows concurrent readers.
	db.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("pragma %q: %w", p, err)
		}
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("schema init: %w", err)
		}
	}

	return &SQLiteClient{db: db}, nil
}

func (c *SQLiteClient) Close() error {
	return c.db.Close()
}

func (c *SQLiteClient) UpsertSession(ctx context.Context, sessionID string) error {
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO sessions(id) VALUES(?) ON CONFLICT(id) DO UPDATE SET last_active=datetime('now')`,
		sessionID,
	)
	return err
}

func (c *SQLiteClient) InsertEvent(ctx context.Context, e Event) error {
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO events(session_id, event_type, role, content, metadata) VALUES(?,?,?,?,?)`,
		e.SessionID, e.EventType, e.Role, e.Content, e.Metadata,
	)
	return err
}

func (c *SQLiteClient) GetSessionEvents(ctx context.Context, sessionID string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT id, session_id, event_type, role, content, metadata, created_at
		 FROM events WHERE session_id=? ORDER BY created_at ASC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (c *SQLiteClient) SearchEvents(ctx context.Context, query string, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT id, session_id, event_type, role, content, metadata, created_at
		 FROM events
		 WHERE content LIKE ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (c *SQLiteClient) ListSessions(ctx context.Context, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT id, created_at, last_active FROM sessions ORDER BY last_active DESC, rowid DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var createdAt, lastActive string
		if err := rows.Scan(&s.ID, &createdAt, &lastActive); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		s.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActive)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (c *SQLiteClient) GetRecentSessionsEvents(ctx context.Context, nSessions, limitPerSession int) ([]Event, error) {
	if nSessions <= 0 {
		nSessions = 3
	}
	if limitPerSession <= 0 {
		limitPerSession = 20
	}

	sessions, err := c.ListSessions(ctx, nSessions)
	if err != nil {
		return nil, err
	}

	var all []Event
	for _, s := range sessions {
		events, err := c.GetSessionEvents(ctx, s.ID, limitPerSession)
		if err != nil {
			return nil, err
		}
		all = append(all, events...)
	}
	return all, nil
}

func scanEvents(rows *sql.Rows) ([]Event, error) {
	var events []Event
	for rows.Next() {
		var e Event
		var createdAt string
		if err := rows.Scan(&e.ID, &e.SessionID, &e.EventType, &e.Role, &e.Content, &e.Metadata, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		events = append(events, e)
	}
	return events, rows.Err()
}

func formatEvents(events []Event) string {
	if len(events) == 0 {
		return "No events found."
	}
	var sb strings.Builder
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("[%s] session=%s role=%s\n%s\n\n",
			e.CreatedAt.Format("2006-01-02 15:04:05"), e.SessionID, e.Role, e.Content))
	}
	return strings.TrimSpace(sb.String())
}
