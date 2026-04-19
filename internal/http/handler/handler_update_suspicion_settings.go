package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) UpdateSuspicionSettings(ctx context.Context, req *gen.SuspicionSettings) (*gen.SuspicionSettings, error) {
	s := suspicionGenToEntity(req)

	out, err := h.sett.UpdateSuspicionSettings(ctx, s)
	if err != nil {
		return nil, err
	}

	return suspicionEntityToGen(out), nil
}
