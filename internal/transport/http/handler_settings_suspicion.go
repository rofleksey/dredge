package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListChannelBlacklist(ctx context.Context) ([]string, error) {
	return h.sett.ListChannelBlacklist(ctx)
}

func (h *Handler) SetChannelBlacklist(ctx context.Context, req *gen.ChannelBlacklistChange) (gen.SetChannelBlacklistRes, error) {
	if err := h.sett.SetChannelBlacklist(ctx, req.Login, req.Add); err != nil {
		return &gen.ErrorMessage{Message: err.Error()}, nil
	}

	return &gen.SetChannelBlacklistNoContent{}, nil
}

func (h *Handler) GetSuspicionSettings(ctx context.Context) (*gen.SuspicionSettings, error) {
	s, err := h.sett.GetSuspicionSettings(ctx)
	if err != nil {
		return nil, err
	}

	return suspicionEntityToGen(s), nil
}

func (h *Handler) UpdateSuspicionSettings(ctx context.Context, req *gen.SuspicionSettings) (*gen.SuspicionSettings, error) {
	s := suspicionGenToEntity(req)

	out, err := h.sett.UpdateSuspicionSettings(ctx, s)
	if err != nil {
		return nil, err
	}

	return suspicionEntityToGen(out), nil
}

func suspicionEntityToGen(s entity.SuspicionSettings) *gen.SuspicionSettings {
	return &gen.SuspicionSettings{
		AutoCheckAccountAge: s.AutoCheckAccountAge,
		AccountAgeSusDays:   s.AccountAgeSusDays,
		AutoCheckBlacklist:  s.AutoCheckBlacklist,
		AutoCheckLowFollows: s.AutoCheckLowFollows,
		LowFollowsThreshold: s.LowFollowsThreshold,
		MaxGqlFollowPages:   s.MaxGQLFollowPages,
	}
}

func suspicionGenToEntity(s *gen.SuspicionSettings) entity.SuspicionSettings {
	if s == nil {
		return entity.SuspicionSettings{}
	}

	return entity.SuspicionSettings{
		AutoCheckAccountAge: s.AutoCheckAccountAge,
		AccountAgeSusDays:   s.AccountAgeSusDays,
		AutoCheckBlacklist:  s.AutoCheckBlacklist,
		AutoCheckLowFollows: s.AutoCheckLowFollows,
		LowFollowsThreshold: s.LowFollowsThreshold,
		MaxGQLFollowPages:   s.MaxGqlFollowPages,
	}
}
