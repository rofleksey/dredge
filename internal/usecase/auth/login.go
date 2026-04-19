package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Usecase) Login(ctx context.Context, email, password string) (string, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.auth.login")
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
