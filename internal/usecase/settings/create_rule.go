package settings

import (
	"context"
	"regexp"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) CreateRule(ctx context.Context, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.create_rule")
	defer span.End()

	if _, err := regexp.Compile(r.Regex); err != nil {
		s.obs.LogError(ctx, span, "compile regex failed", err, zap.String("regex", r.Regex))
		return entity.Rule{}, err
	}

	out, err := s.repo.CreateRule(ctx, r)
	if err != nil {
		s.obs.LogError(ctx, span, "create rule failed", err, zap.String("regex", r.Regex))
	}

	return out, err
}
