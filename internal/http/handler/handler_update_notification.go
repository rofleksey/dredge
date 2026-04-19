package handler

import (
	"context"
	"errors"

	"github.com/go-faster/jx"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

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
