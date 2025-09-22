package controller

import (
	"context"
	"dredge/app/api"
	"dredge/app/mapper"
	"fmt"
	"net/http"

	"github.com/elliotchance/pie/v2"
	"github.com/samber/oops"
)

func (s *Server) SearchMessages(ctx context.Context, req api.SearchMessagesRequestObject) (api.SearchMessagesResponseObject, error) {
	offset := req.Body.Offset
	limit := req.Body.Limit

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	result, err := s.messagesService.Search(ctx, req.Body.Query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("Search: %w", err)
	}

	count, err := s.messagesService.TotalCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("TotalCount: %w", err)
	}

	return api.SearchMessages200JSONResponse{
		Messages:   pie.Map(result, mapper.MapMessage),
		TotalCount: int(count),
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, req api.SendMessageRequestObject) (api.SendMessageResponseObject, error) {
	if !s.limitsService.AllowGlobalRps(ctx, "send_message", 1) {
		return nil, oops.With("statusCode", http.StatusTooManyRequests).New("Too many requests")
	}

	if err := s.accountsService.SendMessage(req.Body.Channel, req.Body.Username, req.Body.Text); err != nil {
		return nil, err
	}

	return api.SendMessage200Response{}, nil
}
