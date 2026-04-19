package auth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Service) ParseToken(ctx context.Context, token string) (int64, string, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.auth.parse_token")
	defer span.End()

	claims := jwt.MapClaims{}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected HMAC variant")
		}

		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		e := fmtInvalidToken()
		s.obs.LogError(ctx, span, "parse token failed", e)
		return 0, "", e
	}

	var id int64

	if sub, ok := claims["sub"].(string); ok {
		id, err = strconv.ParseInt(sub, 10, 64)
		if err != nil {
			e := fmtInvalidToken()
			s.obs.LogError(ctx, span, "parse sub failed", e)
			return 0, "", e
		}
	} else {
		e := fmtInvalidToken()
		s.obs.LogError(ctx, span, "missing sub", e)
		return 0, "", e
	}

	role, _ := claims["role"].(string)
	if role == "" {
		e := fmtInvalidToken()
		s.obs.LogError(ctx, span, "missing role", e)
		return 0, "", e
	}

	return id, role, nil
}
