package auth

import (
	"errors"
	"fmt"
	"time"

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

func fmtInvalidCredentials() error {
	return fmt.Errorf("%w: %w", entity.ErrNoSentry, ErrInvalidCredentials)
}

func fmtInvalidToken() error {
	return fmt.Errorf("%w: invalid token", entity.ErrNoSentry)
}

type Service struct {
	adminEmail        string
	adminPasswordHash string
	jwtSecret         []byte
	jwtTTL            time.Duration
	obs               *observability.Stack
}
