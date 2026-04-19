package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListNotifications(ctx context.Context) ([]gen.NotificationEntry, error) {
	list, err := h.sett.ListNotifications(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]gen.NotificationEntry, 0, len(list))

	for _, e := range list {
		out = append(out, notificationEntityToGen(e))
	}

	return out, nil
}
