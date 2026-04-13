package httptransport

import (
	"context"
	"errors"

	"github.com/go-faster/jx"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
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

func (h *Handler) UpdateNotification(ctx context.Context, req *gen.UpdateNotificationPostRequest) (gen.UpdateNotificationRes, error) {
	var prov *string

	if req.Provider.IsSet() {
		p := string(req.Provider.Value)
		prov = &p
	}

	var settings map[string]any
	if req.Settings.IsSet() {
		settings = rawSettingsToMap(map[string]jx.Raw(req.Settings.Value))
	}

	var enabled *bool

	if req.Enabled.IsSet() {
		v := req.Enabled.Value
		enabled = &v
	}

	e, err := h.sett.UpdateNotification(ctx, req.ID, prov, settings, enabled)
	if err != nil {
		if errors.Is(err, entity.ErrNotificationNotFound) {
			return &gen.ErrorMessage{Message: "notification not found"}, nil
		}
		return nil, err
	}

	out := notificationEntityToGen(e)

	return &out, nil
}

func (h *Handler) DeleteNotification(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteNotificationRes, error) {
	if err := h.sett.DeleteNotification(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrNotificationNotFound) {
			return &gen.ErrorMessage{Message: "notification not found"}, nil
		}
		return nil, err
	}

	return &gen.DeleteNotificationNoContent{}, nil
}
