package entity

import (
	"encoding/json"
	"time"
)

type AISettings struct {
	BaseURL   string
	Model     string
	APIToken  string
	UpdatedAt time.Time
}

type AISettingsPublic struct {
	BaseURL    string
	Model      string
	HasToken   bool
	TokenLast4 string
	UpdatedAt  time.Time
}

type AIConversation struct {
	ID        int64
	Title     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AIMessageRole string

const (
	AIMessageRoleUser      AIMessageRole = "user"
	AIMessageRoleAssistant AIMessageRole = "assistant"
	AIMessageRoleTool      AIMessageRole = "tool"
)

type AIMessage struct {
	ID             int64
	ConversationID int64
	Role           AIMessageRole
	Content        string
	Metadata       map[string]any
	CreatedAt      time.Time
}

func (m *AIMessage) MetadataJSON() ([]byte, error) {
	if m.Metadata == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m.Metadata)
}

func ParseAIMessageMetadata(b []byte) (map[string]any, error) {
	if len(b) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}
