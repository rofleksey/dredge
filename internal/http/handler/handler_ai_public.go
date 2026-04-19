package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

// sanitizeAIMessageForAPI strips tool payloads and internal metadata from messages
// returned to the chat UI. Full content remains in the database for the agent.
func sanitizeAIMessageForAPI(m entity.AIMessage) entity.AIMessage {
	switch m.Role {
	case entity.AIMessageRoleTool:
		return entity.AIMessage{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Role:           m.Role,
			Content:        toolResultPublicContent(m.Content),
			Metadata:       toolMetadataPublic(m.Metadata),
			CreatedAt:      m.CreatedAt,
		}
	case entity.AIMessageRoleAssistant:
		if m.Metadata == nil {
			return m
		}

		meta := make(map[string]any, len(m.Metadata))
		for k, v := range m.Metadata {
			if k == "tool_calls" {
				continue
			}
			meta[k] = v
		}

		return entity.AIMessage{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Role:           m.Role,
			Content:        m.Content,
			Metadata:       meta,
			CreatedAt:      m.CreatedAt,
		}
	default:
		return m
	}
}

func toolMetadataPublic(meta map[string]any) map[string]any {
	if meta == nil {
		return map[string]any{}
	}

	out := make(map[string]any)
	for _, k := range []string{"tool_name", "tool_call_id"} {
		if v, ok := meta[k]; ok {
			out[k] = v
		}
	}

	return out
}

// toolResultPublicContent returns only a textual error from a tool JSON body, or empty on success.
func toolResultPublicContent(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return ""
	}

	errVal, ok := obj["error"]
	if !ok || errVal == nil {
		return ""
	}

	switch e := errVal.(type) {
	case string:
		return e
	case float64:
		return strconv.FormatFloat(e, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(e)
	default:
		b, err := json.Marshal(e)
		if err != nil {
			return ""
		}

		return string(b)
	}
}
