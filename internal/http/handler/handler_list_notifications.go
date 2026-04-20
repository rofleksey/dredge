package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/entity"
)

func (h *Handler) ListNotifications(ctx context.Context, params gen.ListNotificationsParams) ([]gen.NotificationEntry, error) {
	f := entity.NotificationListFilter{}
	if v, ok := params.Limit.Get(); ok {
		f.Limit = v
	}

	if v, ok := params.CursorCreatedAt.Get(); ok {
		f.CursorCreatedAt = &v
	}

	if v, ok := params.CursorID.Get(); ok {
		f.CursorID = &v
	}

	list, err := h.sett.ListNotifications(ctx, f)
	if err != nil {
		return nil, err
	}

	out := make([]gen.NotificationEntry, 0, len(list))

	for _, e := range list {
		out = append(out, notificationEntityToGen(e))
	}

	return out, nil
}
