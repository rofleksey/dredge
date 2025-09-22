package controller

import (
	"context"
	"dredge/app/api"
	"net/http"

	"github.com/samber/oops"
)

func (s *Server) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	if !s.limitsService.AllowIpRpm(ctx, "login", 10) {
		return nil, oops.With("statusCode", http.StatusTooManyRequests).New("Too many requests")
	}

	token, err := s.authService.Login(request.Body.Username, request.Body.Password)
	if err != nil {
		return nil, err
	}

	return api.Login200JSONResponse{Token: token}, nil
}
