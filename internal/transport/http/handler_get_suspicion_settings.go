package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetSuspicionSettings(ctx context.Context) (*gen.SuspicionSettings, error) {
	s, err := h.sett.GetSuspicionSettings(ctx)
	if err != nil {
		return nil, err
	}

	return suspicionEntityToGen(s), nil
}
