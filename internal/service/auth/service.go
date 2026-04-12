package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
)

// Intentional: single built-in admin account id; there is no multi-user account directory.
const fixedAccountID int64 = 1

func New(cfg config.Config, jwtSecret string, jwtTTL time.Duration, obs *observability.Stack) (*Service, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}

	return &Service{
		adminEmail:        cfg.Admin.Email,
		adminPasswordHash: string(hash),
		jwtSecret:         []byte(jwtSecret),
		jwtTTL:            jwtTTL,
		obs:               obs,
	}, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.auth.login")
	defer span.End()

	s.obs.Logger.Debug("auth login attempt")

	if email != s.adminEmail {
		return "", fmtInvalidCredentials()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(s.adminPasswordHash), []byte(password)); err != nil {
		return "", fmtInvalidCredentials()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  strconv.FormatInt(fixedAccountID, 10),
		"role": "admin",
		"exp":  time.Now().Add(s.jwtTTL).Unix(),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.obs.LogError(ctx, span, "sign jwt failed", err)
		return "", err
	}

	return signed, nil
}

func fmtInvalidCredentials() error {
	return fmt.Errorf("%w: %w", entity.ErrNoSentry, ErrInvalidCredentials)
}

func (s *Service) ParseToken(ctx context.Context, token string) (int64, string, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.auth.parse_token")
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

func fmtInvalidToken() error {
	return fmt.Errorf("%w: invalid token", entity.ErrNoSentry)
}

func (s *Service) Me(ctx context.Context, accountID int64) (entity.Account, error) {
	_, span := s.obs.StartSpan(ctx, "service.auth.me")
	defer span.End()

	if accountID != fixedAccountID {
		return entity.Account{}, fmtInvalidToken()
	}

	return entity.Account{
		ID:    fixedAccountID,
		Email: s.adminEmail,
		Role:  "admin",
	}, nil
}

type Service struct {
	adminEmail        string
	adminPasswordHash string
	jwtSecret         []byte
	jwtTTL            time.Duration
	obs               *observability.Stack
}
