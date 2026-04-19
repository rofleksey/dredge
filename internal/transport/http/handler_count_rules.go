package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) CountRules(ctx context.Context) (*gen.CountResponse, error) {
	n, err := h.sett.CountRules(ctx)
	if err != nil {
		return nil, err
	}

	return &gen.CountResponse{Total: n}, nil
}
