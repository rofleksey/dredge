package auth

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) Me(ctx context.Context, accountID int64) (entity.Account, error) {
	_, span := s.obs.StartSpan(ctx, "usecase.auth.me")
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
