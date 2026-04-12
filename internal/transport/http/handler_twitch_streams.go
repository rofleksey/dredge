package httptransport

import (
	"context"
	"errors"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListRecordedStreams(ctx context.Context, params gen.ListRecordedStreamsParams) ([]gen.RecordedStream, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_streams")
	defer span.End()

	f := entity.StreamListFilter{}

	if params.Limit.IsSet() {
		f.Limit = params.Limit.Value
	}

	if params.ChannelLogin.IsSet() {
		f.ChannelLogin = params.ChannelLogin.Value
	}

	if ct, ok1 := params.CursorStartedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			f.CursorStartedAt = &t
			i := id
			f.CursorID = &i
		}
	}

	list, err := h.twitch.ListMonitoredStreams(ctx, f)
	if err != nil {
		h.obs.LogError(ctx, span, "list recorded streams failed", err)
		return nil, err
	}

	out := make([]gen.RecordedStream, 0, len(list))
	for _, s := range list {
		out = append(out, streamEntityToGen(s))
	}

	return out, nil
}

func (h *Handler) GetRecordedStream(ctx context.Context, params gen.GetRecordedStreamParams) (gen.GetRecordedStreamRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_recorded_stream")
	defer span.End()

	st, err := h.twitch.GetMonitoredStream(ctx, params.StreamId)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "get recorded stream failed", err)
		return nil, err
	}

	g := streamEntityToGen(st)

	return &g, nil
}

func (h *Handler) ListRecordedStreamMessages(ctx context.Context, params gen.ListRecordedStreamMessagesParams) (gen.ListRecordedStreamMessagesRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_stream_messages")
	defer span.End()

	limit := 50
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 200 {
		limit = 200
	}

	f := entity.ChatMessageListFilter{Limit: limit}

	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	if params.Text.IsSet() {
		f.Text = params.Text.Value
	}

	if params.ChatterUserID.IsSet() {
		v := params.ChatterUserID.Value
		f.ChatterUserID = &v
	}

	if ct, ok1 := params.CursorCreatedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			f.CursorCreatedAt = &t
			i := id
			f.CursorID = &i
		}
	}

	list, err := h.twitch.ListStreamMessages(ctx, params.StreamId, f)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "list recorded stream messages failed", err)
		return nil, err
	}

	out := make(gen.ListRecordedStreamMessagesOKApplicationJSON, 0, len(list))
	for _, m := range list {
		out = append(out, chatHistoryEntityToGen(m))
	}

	return &out, nil
}

func (h *Handler) ListRecordedStreamActivity(ctx context.Context, params gen.ListRecordedStreamActivityParams) (gen.ListRecordedStreamActivityRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_stream_activity")
	defer span.End()

	limit := 50
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 200 {
		limit = 200
	}

	var (
		cursorCreatedAt *time.Time
		cursorID        *int64
	)

	if ct, ok1 := params.CursorCreatedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			cursorCreatedAt = &t
			i := id
			cursorID = &i
		}
	}

	list, err := h.twitch.ListStreamActivity(ctx, params.StreamId, limit, cursorCreatedAt, cursorID)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "list recorded stream activity failed", err)
		return nil, err
	}

	out := make(gen.ListRecordedStreamActivityOKApplicationJSON, 0, len(list))
	for _, e := range list {
		out = append(out, entityActivityToGen(e, ""))
	}

	return &out, nil
}

func (h *Handler) GetRecordedStreamLeaderboard(ctx context.Context, params gen.GetRecordedStreamLeaderboardParams) (gen.GetRecordedStreamLeaderboardRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_recorded_stream_leaderboard")
	defer span.End()

	st, err := h.twitch.GetMonitoredStream(ctx, params.StreamId)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "get stream for leaderboard failed", err)
		return nil, err
	}

	sort := entity.StreamLeaderboardSortPresenceDesc
	if v, ok := params.Sort.Get(); ok {
		sort = entity.StreamLeaderboardSort(v)
	}

	q := ""
	if v, ok := params.Q.Get(); ok {
		q = v
	}

	rows, err := h.twitch.StreamLeaderboard(ctx, st, sort, q)
	if err != nil {
		h.obs.LogError(ctx, span, "stream leaderboard failed", err)
		return nil, err
	}

	out := make(gen.GetRecordedStreamLeaderboardOKApplicationJSON, 0, len(rows))
	for _, r := range rows {
		row := gen.StreamLeaderboardEntry{
			Login:           r.Login,
			UserTwitchID:    r.UserTwitchID,
			PresenceSeconds: r.PresenceSeconds,
			MessageCount:    r.MessageCount,
		}

		if r.AccountCreatedAt != nil {
			row.AccountCreatedAt = gen.NewOptNilDateTime(*r.AccountCreatedAt)
		} else {
			var z gen.OptNilDateTime
			z.SetToNull()
			row.AccountCreatedAt = z
		}

		out = append(out, row)
	}

	return &out, nil
}
