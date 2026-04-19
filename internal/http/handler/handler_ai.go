package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-faster/jx"
	"github.com/jackc/pgx/v5"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) GetAiSettings(ctx context.Context) (*gen.AiSettings, error) {
	_, pub, err := h.ai.GetAISettings(ctx)
	if err != nil {
		return nil, err
	}
	return genAiSettings(pub), nil
}

func (h *Handler) PatchAiSettings(ctx context.Context, req *gen.PatchAiSettingsRequest) (*gen.AiSettings, error) {
	var baseURL, model, tok *string
	if req.BaseURL.IsSet() {
		v := req.BaseURL.Value
		baseURL = &v
	}
	if req.Model.IsSet() {
		v := req.Model.Value
		model = &v
	}
	if req.APIToken.IsSet() {
		v := req.APIToken.Value
		tok = &v
	}
	pub, err := h.ai.PatchAISettings(ctx, baseURL, model, tok)
	if err != nil {
		return nil, err
	}
	return genAiSettings(pub), nil
}

func genAiSettings(pub entity.AISettingsPublic) *gen.AiSettings {
	var tl4 gen.OptString
	if pub.HasToken && pub.TokenLast4 != "" {
		tl4 = gen.NewOptString(pub.TokenLast4)
	}
	return &gen.AiSettings{
		BaseURL:    pub.BaseURL,
		Model:      pub.Model,
		HasToken:   pub.HasToken,
		TokenLast4: tl4,
		UpdatedAt:  pub.UpdatedAt,
	}
}

func (h *Handler) ListAiConversations(ctx context.Context) ([]gen.AiConversation, error) {
	list, err := h.ai.ListAIConversations(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]gen.AiConversation, 0, len(list))
	for _, c := range list {
		out = append(out, genAiConversation(c))
	}
	return out, nil
}

func genAiConversation(c entity.AIConversation) gen.AiConversation {
	var title gen.OptNilString
	if c.Title != nil {
		title.SetTo(*c.Title)
	} else {
		title.SetToNull()
	}
	return gen.AiConversation{
		ID:        c.ID,
		Title:     title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (h *Handler) CreateAiConversation(ctx context.Context, req gen.OptCreateAiConversationRequest) (*gen.AiConversation, error) {
	var title *string
	if req.IsSet() && req.Value.Title.IsSet() {
		t := req.Value.Title.Value
		title = &t
	}
	c, err := h.ai.CreateAIConversation(ctx, title)
	if err != nil {
		return nil, err
	}
	g := genAiConversation(c)
	return &g, nil
}

func (h *Handler) DeleteAiConversation(ctx context.Context, params gen.DeleteAiConversationParams) (gen.DeleteAiConversationRes, error) {
	err := h.ai.DeleteAIConversation(ctx, params.ConversationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &gen.ErrorMessage{Message: "conversation not found"}, nil
		}
		return nil, err
	}
	return &gen.DeleteAiConversationNoContent{}, nil
}

func (h *Handler) ListAiMessages(ctx context.Context, params gen.ListAiMessagesParams) (gen.ListAiMessagesRes, error) {
	_, err := h.ai.GetAIConversation(ctx, params.ConversationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &gen.ErrorMessage{Message: "conversation not found"}, nil
		}
		return nil, err
	}
	msgs, err := h.ai.ListAIMessages(ctx, params.ConversationId)
	if err != nil {
		return nil, err
	}
	out := make(gen.ListAiMessagesOKApplicationJSON, 0, len(msgs))
	for _, m := range msgs {
		sm := sanitizeAIMessageForAPI(m)
		if sm.Role == entity.AIMessageRoleTool && strings.TrimSpace(sm.Content) == "" {
			continue
		}

		out = append(out, genAiMessage(sm))
	}
	return &out, nil
}

func genAiMessage(m entity.AIMessage) gen.AiMessage {
	meta := make(gen.AiMessageMetadata)
	for k, v := range m.Metadata {
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		meta[k] = jx.Raw(b)
	}
	return gen.AiMessage{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Role:           gen.AiMessageRole(m.Role),
		Content:        m.Content,
		Metadata:       meta,
		CreatedAt:      m.CreatedAt,
	}
}

func (h *Handler) CreateAiMessage(ctx context.Context, req *gen.CreateAiMessageRequest, params gen.CreateAiMessageParams) (gen.CreateAiMessageRes, error) {
	_, err := h.ai.GetAIConversation(ctx, params.ConversationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &gen.ErrorMessage{Message: "conversation not found"}, nil
		}
		return nil, err
	}
	if err := h.ai.StartAgentRun(params.ConversationId, req.GetContent()); err != nil {
		return nil, err
	}
	return &gen.AiRunAccepted{Accepted: true}, nil
}

func (h *Handler) ConfirmAiTool(ctx context.Context, req *gen.ConfirmAiToolRequest, params gen.ConfirmAiToolParams) (gen.ConfirmAiToolRes, error) {
	if err := h.ai.ConfirmPendingTool(params.ConversationId, req.GetToolCallID(), req.GetApprove()); err != nil {
		return &gen.ErrorMessage{Message: err.Error()}, nil
	}
	return &gen.AiRunAccepted{Accepted: true}, nil
}

func (h *Handler) StopAiAgent(ctx context.Context, params gen.StopAiAgentParams) (gen.StopAiAgentRes, error) {
	_, err := h.ai.GetAIConversation(ctx, params.ConversationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &gen.ErrorMessage{Message: "conversation not found"}, nil
		}
		return nil, err
	}
	h.ai.StopRun(params.ConversationId)
	return &gen.StopAiAgentNoContent{}, nil
}
