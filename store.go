package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// UserPromptPayload is the JSON Claude Code sends on UserPromptSubmit.
type UserPromptPayload struct {
	SessionID string `json:"session_id"`
	Prompt    string `json:"prompt"`
}

// StopPayload is the JSON Claude Code sends on Stop.
type StopPayload struct {
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
	Reasoning string `json:"reasoning"`
	Response  string `json:"response"`
}

func RunStore(eventType string) error {
	return storeFromReader(os.Stdin, eventType)
}

func storeFromReader(r io.Reader, eventType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	db, err := NewSQLiteClient()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	switch eventType {
	case "UserPromptSubmit":
		var p UserPromptPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("parse payload: %w", err)
		}
		if p.SessionID == "" || p.Prompt == "" {
			return nil
		}
		if err := db.UpsertSession(ctx, p.SessionID); err != nil {
			return fmt.Errorf("upsert session: %w", err)
		}
		return db.InsertEvent(ctx, Event{
			SessionID: p.SessionID,
			EventType: eventType,
			Role:      "user",
			Content:   p.Prompt,
			Metadata:  "{}",
		})

	case "Stop":
		var p StopPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("parse payload: %w", err)
		}
		if p.SessionID == "" || p.Response == "" {
			return nil
		}
		if err := db.UpsertSession(ctx, p.SessionID); err != nil {
			return fmt.Errorf("upsert session: %w", err)
		}
		return db.InsertEvent(ctx, Event{
			SessionID: p.SessionID,
			EventType: eventType,
			Role:      "assistant",
			Content:   p.Response,
			Metadata:  "{}",
		})

	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}
