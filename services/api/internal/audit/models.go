package audit

import (
	"encoding/json"
	"time"
)

type LogItem struct {
	ID             int64           `json:"id"`
	ActorUserID    *string         `json:"actor_user_id"`
	ActorName      *string         `json:"actor_name"`
	ActorRoleCodes []string        `json:"actor_role_codes"`
	Action         string          `json:"action"`
	EntityType     string          `json:"entity_type"`
	EntityID       *string         `json:"entity_id"`
	Metadata       json.RawMessage `json:"metadata"`
	BeforeData     json.RawMessage `json:"before_data"`
	AfterData      json.RawMessage `json:"after_data"`
	RequestID      *string         `json:"request_id"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Filter struct {
	Action      string
	ActorUserID string
	EntityType  string
	EntityID    string
	Limit       int
	Cursor      int64
}

type ListResult struct {
	Data []LogItem `json:"data"`
	Meta Meta      `json:"meta"`
}

type Meta struct {
	NextCursor *int64 `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}