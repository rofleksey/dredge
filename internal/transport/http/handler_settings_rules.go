package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListRules(ctx context.Context) ([]gen.Rule, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_rules")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	list, err := h.sett.ListRules(ctx)
	if err != nil {
		h.obs.LogError(ctx, span, "list rules failed", err)
		return nil, err
	}

	out := make([]gen.Rule, 0, len(list))
	for _, r := range list {
		out = append(out, ruleEntityToGen(r))
	}

	return out, nil
}

func (h *Handler) CountRules(ctx context.Context) (*gen.CountResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	n, err := h.sett.CountRules(ctx)
	if err != nil {
		return nil, err
	}
	return &gen.CountResponse{Total: n}, nil
}

func (h *Handler) CreateRule(ctx context.Context, req *gen.CreateRuleRequest) (*gen.Rule, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_rule")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	r, err := h.sett.CreateRule(ctx, createRuleReqToEntity(req))
	if err != nil {
		h.obs.LogError(ctx, span, "create rule failed", err, zap.String("regex", req.Regex))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.String("regex", req.Regex))
		return nil, err
	}

	out := ruleEntityToGen(r)

	return &out, nil
}

func (h *Handler) UpdateRule(ctx context.Context, req *gen.UpdateRulePostRequest) (gen.UpdateRuleRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_rule")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	r, err := h.sett.UpdateRule(ctx, req.GetID(), updateRulePostReqToEntity(req))
	if err != nil {
		if errors.Is(err, entity.ErrRuleNotFound) {
			return &gen.ErrorMessage{Message: "rule not found"}, nil
		}

		h.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	out := ruleEntityToGen(r)

	return &out, nil
}

func (h *Handler) DeleteRule(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteRuleRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.delete_rule")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if err := h.sett.DeleteRule(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrRuleNotFound) {
			return &gen.ErrorMessage{Message: "rule not found"}, nil
		}

		h.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	return &gen.DeleteRuleNoContent{}, nil
}
