package handler

import (
	"context"

	"github.com/go-faster/jx"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) CreateNotification(ctx context.Context, req *gen.CreateNotificationRequest) (*gen.NotificationEntry, error) {
	enabled := true
	if req.Enabled.IsSet() {
		enabled = req.Enabled.Value
	}

	e, err := h.sett.CreateNotification(ctx, string(req.Provider), rawSettingsToMap(map[string]jx.Raw(req.Settings)), enabled)
	if err != nil {
		return nil, err
	}

	out := notificationEntityToGen(e)

	return &out, nil
}
