package ai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	agentSystemPrompt = `You are the Dredge admin assistant for a single-operator Twitch monitoring app.
Use the provided tools to inspect persisted chat, user activity, and automation rules.
Prefer read-only tools. Mutating tools require explicit user confirmation in the UI before they run.
Before creating or changing rules, call list_rules and rule_template_variables so you use valid event_type, action_type, middleware types, and message placeholders.
When creating interval rules, use event_type "interval" with event_settings containing interval_seconds (positive integer) and channel (login).`

	maxAgentIterations = 24

	defaultChatMaxTokens   = 8192
	defaultChatTemperature = float32(1.0)
	emptyCompletionRetries = 3
)

func (u *Usecase) waitRunStopped(convID int64, max time.Duration) {
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		if _, ok := u.runs.Load(convID); !ok {
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
}

// StartAgentRun inserts the user message and starts the agent loop asynchronously.
func (u *Usecase) StartAgentRun(convID int64, userText string) error {
	if strings.TrimSpace(userText) == "" {
		return errors.New("empty message")
	}

	u.StopRun(convID)
	u.waitRunStopped(convID, 3*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	st := &runState{
		cancel:    cancel,
		approveCh: make(chan bool, 1),
	}
	u.runs.Store(convID, st)

	_, err := u.repo.InsertAIMessage(ctx, entity.AIMessage{
		ConversationID: convID,
		Role:           entity.AIMessageRoleUser,
		Content:        userText,
		Metadata:       map[string]any{},
	})
	if err != nil {
		u.runs.Delete(convID)
		cancel()
		return err
	}
	_ = u.repo.TouchAIConversation(ctx, convID)

	go func() {
		defer u.runs.Delete(convID)
		defer cancel()

		u.agentLoop(ctx, convID, st)
	}()

	return nil
}

func (u *Usecase) broadcast(convID int64, kind string, payload map[string]any) {
	if u.hub == nil {
		return
	}
	m := map[string]any{
		"type":            "ai_agent",
		"conversation_id": convID,
		"kind":            kind,
		"ts":              time.Now().UTC().Format(time.RFC3339Nano),
	}
	for k, v := range payload {
		m[k] = v
	}
	u.hub.BroadcastJSON(m)
}

func (u *Usecase) agentLoop(ctx context.Context, convID int64, st *runState) {
	settings, err := u.repo.GetAISettings(ctx)
	if err != nil {
		u.agentErr(ctx, convID, err, "load settings")
		return
	}
	token := settings.APIToken
	if strings.TrimSpace(settings.BaseURL) == "" || strings.TrimSpace(token) == "" {
		u.agentErr(ctx, convID, errors.New("configure AI base URL and API token in Settings"), "config")
		return
	}

	cfg := openai.DefaultConfig(token)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	client := openai.NewClientWithConfig(cfg)
	model := strings.TrimSpace(settings.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}

	msgs, err := u.openAIMessagesFromDB(ctx, convID)
	if err != nil {
		u.agentErr(ctx, convID, err, "load messages")
		return
	}

	tools := LLMTools()

iterLoop:
	for iter := 0; iter < maxAgentIterations; iter++ {
		select {
		case <-ctx.Done():
			u.persistSystemNote(ctx, convID, "Agent stopped.")
			u.broadcast(convID, "done", map[string]any{"reason": "stopped"})
			return
		default:
		}

		req := openai.ChatCompletionRequest{
			Model:       model,
			Messages:    msgs,
			Tools:       tools,
			MaxTokens:   defaultChatMaxTokens,
			Temperature: defaultChatTemperature,
		}

		u.broadcast(convID, "llm_request", map[string]any{"iteration": iter})

		var resp openai.ChatCompletionResponse
		var err error
		for attempt := 0; attempt < emptyCompletionRetries; attempt++ {
			resp, err = client.CreateChatCompletion(ctx, req)
			if err != nil {
				u.agentErr(ctx, convID, err, "llm request")
				return
			}
			if len(resp.Choices) > 0 {
				break
			}
			if u.obs != nil && u.obs.Logger != nil {
				u.obs.Logger.Warn("ai chat completion returned no choices",
					zap.String("model", model),
					zap.Int("agent_iteration", iter),
					zap.Int("retry", attempt),
					zap.String("response_id", resp.ID),
				)
			}
			if attempt < emptyCompletionRetries-1 {
				time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
			}
		}
		if len(resp.Choices) == 0 {
			u.agentErr(ctx, convID, errors.New("empty completion after retries (check model/provider compatibility with tools)"), "llm")
			return
		}

		choice := resp.Choices[0].Message

		if len(choice.ToolCalls) == 0 {
			txt := choice.Content
			_, err := u.repo.InsertAIMessage(ctx, entity.AIMessage{
				ConversationID: convID,
				Role:           entity.AIMessageRoleAssistant,
				Content:        txt,
				Metadata:       map[string]any{},
			})
			if err != nil {
				u.agentErr(ctx, convID, err, "save assistant")
				return
			}
			_ = u.repo.TouchAIConversation(ctx, convID)
			u.broadcast(convID, "message", map[string]any{"role": "assistant", "content": txt})
			u.broadcast(convID, "done", map[string]any{"reason": "completed"})
			return
		}

		toolCallsJSON, _ := json.Marshal(choice.ToolCalls)
		_, err = u.repo.InsertAIMessage(ctx, entity.AIMessage{
			ConversationID: convID,
			Role:           entity.AIMessageRoleAssistant,
			Content:        choice.Content,
			Metadata: map[string]any{
				"tool_calls": string(toolCallsJSON),
			},
		})
		if err != nil {
			u.agentErr(ctx, convID, err, "save assistant tools")
			return
		}
		_ = u.repo.TouchAIConversation(ctx, convID)
		u.broadcast(convID, "message", map[string]any{"role": "assistant", "has_tool_calls": true})

		msgs = append(msgs, choice)

		for _, tc := range choice.ToolCalls {
			name := tc.Function.Name
			args := tc.Function.Arguments

			u.broadcast(convID, "tool_attempt", map[string]any{
				"tool_name":    name,
				"tool_call_id": tc.ID,
			})

			if ToolIsReadOnly(name) {
				resText, execErr := u.ExecTool(ctx, name, args)
				if execErr != nil {
					resText = mustJSON(map[string]string{"error": execErr.Error()})
				}
				if _, err := u.repo.InsertAIMessage(ctx, entity.AIMessage{
					ConversationID: convID,
					Role:           entity.AIMessageRoleTool,
					Content:        resText,
					Metadata: map[string]any{
						"tool_call_id": tc.ID,
						"tool_name":    name,
					},
				}); err != nil {
					u.agentErr(ctx, convID, err, "save tool result")
					return
				}
				_ = u.repo.TouchAIConversation(ctx, convID)
				msgs = append(msgs, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    resText,
					ToolCallID: tc.ID,
				})
				u.broadcast(convID, "tool_result", map[string]any{"tool_name": name, "tool_call_id": tc.ID, "ok": execErr == nil})
				continue
			}

			u.broadcast(convID, "needs_confirmation", map[string]any{
				"tool_name":    name,
				"tool_call_id": tc.ID,
			})

			st.mu.Lock()
			st.pendingID = tc.ID
			st.mu.Unlock()

			var approve bool
			select {
			case <-ctx.Done():
				u.persistSystemNote(ctx, convID, "Agent stopped while waiting for confirmation.")
				u.broadcast(convID, "done", map[string]any{"reason": "stopped"})
				return
			case approve = <-st.approveCh:
			}

			st.mu.Lock()
			st.pendingID = ""
			st.mu.Unlock()

			var resText string
			if !approve {
				resText = `{"error":"user rejected this tool call"}`
			} else {
				var execErr error
				resText, execErr = u.ExecTool(ctx, name, args)
				if execErr != nil {
					resText = mustJSON(map[string]string{"error": execErr.Error()})
				}
			}

			if _, err := u.repo.InsertAIMessage(ctx, entity.AIMessage{
				ConversationID: convID,
				Role:           entity.AIMessageRoleTool,
				Content:        resText,
				Metadata: map[string]any{
					"tool_call_id": tc.ID,
					"tool_name":    name,
				},
			}); err != nil {
				u.agentErr(ctx, convID, err, "save tool result")
				return
			}
			_ = u.repo.TouchAIConversation(ctx, convID)
			msgs = append(msgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    resText,
				ToolCallID: tc.ID,
			})
			u.broadcast(convID, "tool_result", map[string]any{"tool_name": name, "tool_call_id": tc.ID, "approved": approve})
		}

		continue iterLoop
	}

	u.agentErr(ctx, convID, errors.New("max agent iterations"), "limit")
}

func (u *Usecase) persistSystemNote(ctx context.Context, convID int64, text string) {
	_, err := u.repo.InsertAIMessage(ctx, entity.AIMessage{
		ConversationID: convID,
		Role:           entity.AIMessageRoleAssistant,
		Content:        text,
		Metadata:       map[string]any{"system": true},
	})
	if err != nil && u.obs != nil && u.obs.Logger != nil {
		u.obs.Logger.Debug("ai persist note failed", zap.Error(err))
	}
	_ = u.repo.TouchAIConversation(ctx, convID)
}

func (u *Usecase) agentErr(ctx context.Context, convID int64, err error, phase string) {
	if u.obs != nil && u.obs.Logger != nil {
		u.obs.Logger.Warn("ai agent error", zap.String("phase", phase), zap.Error(err))
	}
	msg := err.Error()
	_, _ = u.repo.InsertAIMessage(ctx, entity.AIMessage{
		ConversationID: convID,
		Role:           entity.AIMessageRoleAssistant,
		Content:        "Error: " + msg,
		Metadata:       map[string]any{"error": true, "phase": phase},
	})
	_ = u.repo.TouchAIConversation(ctx, convID)
	u.broadcast(convID, "error", map[string]any{"message": msg, "phase": phase})
	u.broadcast(convID, "done", map[string]any{"reason": "error"})
}

func (u *Usecase) openAIMessagesFromDB(ctx context.Context, convID int64) ([]openai.ChatCompletionMessage, error) {
	rows, err := u.repo.ListAIMessages(ctx, convID)
	if err != nil {
		return nil, err
	}
	out := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: agentSystemPrompt},
	}
	for _, m := range rows {
		switch m.Role {
		case entity.AIMessageRoleUser:
			out = append(out, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: m.Content})
		case entity.AIMessageRoleAssistant:
			cm := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: m.Content}
			if raw, ok := m.Metadata["tool_calls"]; ok {
				var tcs []openai.ToolCall
				switch v := raw.(type) {
				case string:
					_ = json.Unmarshal([]byte(v), &tcs)
				case json.RawMessage:
					_ = json.Unmarshal(v, &tcs)
				case []byte:
					_ = json.Unmarshal(v, &tcs)
				case []any:
					b, _ := json.Marshal(v)
					_ = json.Unmarshal(b, &tcs)
				}
				if len(tcs) > 0 {
					cm.ToolCalls = tcs
				}
			}
			out = append(out, cm)
		case entity.AIMessageRoleTool:
			tid, _ := m.Metadata["tool_call_id"].(string)
			out = append(out, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    m.Content,
				ToolCallID: tid,
			})
		}
	}
	return out, nil
}

// ConfirmPendingTool resumes a blocked mutating tool (called from HTTP handler).
func (u *Usecase) ConfirmPendingTool(convID int64, toolCallID string, approve bool) error {
	v, ok := u.runs.Load(convID)
	if !ok {
		return errors.New("no active run for conversation")
	}
	st := v.(*runState)
	st.mu.Lock()
	match := st.pendingID == toolCallID
	st.mu.Unlock()
	if !match {
		return errors.New("no pending tool call with that id")
	}
	select {
	case st.approveCh <- approve:
	default:
		return errors.New("confirmation already sent")
	}
	return nil
}
