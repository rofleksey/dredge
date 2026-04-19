package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) DeleteNotification(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteNotificationRes, error) {
	if err := h.sett.DeleteNotification(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrNotificationNotFound) {
			return &gen.ErrorMessage{Message: "notification not found"}, nil
		}

		return nil, err
	}

	return &gen.DeleteNotificationNoContent{}, nil
}
