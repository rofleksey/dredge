package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) CountRules(ctx context.Context) (*gen.CountResponse, error) {
	n, err := h.sett.CountRules(ctx)
	if err != nil {
		return nil, err
	}

	return &gen.CountResponse{Total: n}, nil
}
