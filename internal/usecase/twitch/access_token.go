package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// accessTokenForFirstLinkedAccount returns a cached OAuth access token (via helix.Client.CachedUserAccessTokenForAccount, ~30m)
// for the first linked Twitch account, persisting a rotated refresh token when Twitch returns one.
func (s *Service) accessTokenForFirstLinkedAccount(ctx context.Context) (accessToken string, acc entity.TwitchAccount, err error) {
	accs, err := s.repo.ListTwitchAccounts(ctx)
	if err != nil {
		return "", entity.TwitchAccount{}, err
	}

	if len(accs) == 0 {
		return "", entity.TwitchAccount{}, ErrNoLinkedTwitchAccount
	}

	// List omits refresh_token; load full row for OAuth.
	acc, err = s.repo.GetTwitchAccountByID(ctx, accs[0].ID)
	if err != nil {
		return "", entity.TwitchAccount{}, err
	}

	at, newRT, err := s.CachedUserAccessTokenForAccount(ctx, acc.ID, acc.RefreshToken)
	if err != nil {
		return "", acc, err
	}

	if newRT != "" && newRT != acc.RefreshToken {
		_ = s.repo.UpdateTwitchRefreshToken(ctx, acc.ID, newRT)
	}

	return at, acc, nil
}
