package httptransport

import (
	"context"
)

func (h *Handler) ListChannelBlacklist(ctx context.Context) ([]string, error) {
	return h.sett.ListChannelBlacklist(ctx)
}
