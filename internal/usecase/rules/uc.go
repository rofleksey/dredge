package rules

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
)

// Usecase persists rules and refreshes the engine snapshot.
type Usecase struct {
	repo    repository.Store
	obs     *observability.Stack
	engine  *Engine
	restart func(context.Context) error
}

// NewUsecase constructs the rules use case. restart is called after successful writes (e.g. IRC reconnect).
func NewUsecase(repo repository.Store, obs *observability.Stack, engine *Engine, restart func(context.Context) error) *Usecase {
	return &Usecase{
		repo:    repo,
		obs:     obs,
		engine:  engine,
		restart: restart,
	}
}

// Engine returns the rule engine (for tests and wiring).
func (s *Usecase) Engine() *Engine {
	return s.engine
}

func (s *Usecase) reloadEngine(ctx context.Context) error {
	if s.engine == nil {
		return nil
	}

	list, err := s.repo.ListRules(ctx)
	if err != nil {
		return err
	}

	s.engine.Reload(ctx, list)

	return nil
}

func (s *Usecase) triggerRestart(ctx context.Context) {
	if s.restart == nil {
		return
	}

	if err := s.restart(ctx); err != nil {
		s.obs.Logger.Warn("rules restart hook failed", zap.Error(err))
	}
}

// Bootstrap loads rules into the engine on process start.
func (s *Usecase) Bootstrap(ctx context.Context) error {
	return s.reloadEngine(ctx)
}

func (s *Usecase) ListRules(ctx context.Context) ([]entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.list_rules")
	defer span.End()

	out, err := s.repo.ListRules(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list rules failed", err)
	}

	return out, err
}

func (s *Usecase) CountRules(ctx context.Context) (int64, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.count_rules")
	defer span.End()

	n, err := s.repo.CountRules(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "count rules failed", err)
	}

	return n, err
}

func (s *Usecase) CreateRule(ctx context.Context, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.create_rule")
	defer span.End()

	if err := ValidateRule(r); err != nil {
		s.obs.LogError(ctx, span, "validate rule failed", err, zap.String("event_type", r.EventType))
		return entity.Rule{}, err
	}

	if err := s.validateSendChatAccount(ctx, r); err != nil {
		s.obs.LogError(ctx, span, "validate rule failed", err, zap.String("event_type", r.EventType))
		return entity.Rule{}, err
	}

	out, err := s.repo.CreateRule(ctx, r)
	if err != nil {
		s.obs.LogError(ctx, span, "create rule failed", err, zap.String("event_type", r.EventType))
		return entity.Rule{}, err
	}

	if err := s.reloadEngine(ctx); err != nil {
		s.obs.LogError(ctx, span, "reload engine after create failed", err)
	}

	s.triggerRestart(ctx)

	return out, nil
}

func (s *Usecase) UpdateRule(ctx context.Context, id int64, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.update_rule")
	defer span.End()

	if err := ValidateRule(r); err != nil {
		s.obs.LogError(ctx, span, "validate rule failed", err, zap.Int64("id", id))
		return entity.Rule{}, err
	}

	if err := s.validateSendChatAccount(ctx, r); err != nil {
		s.obs.LogError(ctx, span, "validate rule failed", err, zap.Int64("id", id))
		return entity.Rule{}, err
	}

	out, err := s.repo.UpdateRule(ctx, id, r)
	if err != nil {
		if !errors.Is(err, entity.ErrRuleNotFound) {
			s.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", id))
		}
		return entity.Rule{}, err
	}

	if err := s.reloadEngine(ctx); err != nil {
		s.obs.LogError(ctx, span, "reload engine after update failed", err)
	}

	s.triggerRestart(ctx)

	return out, nil
}

func (s *Usecase) DeleteRule(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.delete_rule")
	defer span.End()

	err := s.repo.DeleteRule(ctx, id)
	if err != nil {
		if !errors.Is(err, entity.ErrRuleNotFound) {
			s.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", id))
		}
		return err
	}

	if err := s.reloadEngine(ctx); err != nil {
		s.obs.LogError(ctx, span, "reload engine after delete failed", err)
	}

	s.triggerRestart(ctx)

	return nil
}

func (s *Usecase) validateSendChatAccount(ctx context.Context, r entity.Rule) error {
	if r.ActionType != ActionSendChat {
		return nil
	}

	aid, err := ParseSendChatAccountID(r.ActionSettings)
	if err != nil {
		return err
	}

	if aid <= 0 {
		return nil
	}

	_, err = s.repo.GetTwitchAccountByID(ctx, aid)
	if err == nil {
		return nil
	}

	if errors.Is(err, entity.ErrTwitchAccountNotFound) {
		return fmt.Errorf("send_chat: Twitch account is not linked in this app: %w", entity.ErrInvalidRule)
	}

	return err
}
