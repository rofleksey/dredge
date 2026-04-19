package settings

import (
	"context"
	"regexp"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) UpdateRule(ctx context.Context, id int64, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.update_rule")
	defer span.End()

	if _, err := regexp.Compile(r.Regex); err != nil {
		s.obs.LogError(ctx, span, "compile regex failed", err, zap.String("regex", r.Regex))
		return entity.Rule{}, err
	}

	out, err := s.repo.UpdateRule(ctx, id, r)
	if err != nil {
		s.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", id))
	}

	return out, err
}
