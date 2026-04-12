package twitch

import (
	"testing"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestUserIDsToEnrichMissingAccounts_orderAndCap(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)

	list := []entity.ChannelChatterEntry{
		{Login: "a", UserTwitchID: 1, PresentSince: t1, AccountCreatedAt: nil},
		{Login: "b", UserTwitchID: 2, PresentSince: t3, AccountCreatedAt: nil},
		{Login: "c", UserTwitchID: 3, PresentSince: t2, AccountCreatedAt: nil},
		{Login: "d", UserTwitchID: 4, PresentSince: t2, AccountCreatedAt: ptrTime(t1)},
	}

	ids := userIDsToEnrichMissingAccounts(list, 2)
	assert.Equal(t, []int64{2, 3}, ids)
}

func TestUserIDsToEnrichMissingAccounts_allHaveAccount(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	list := []entity.ChannelChatterEntry{
		{Login: "a", UserTwitchID: 1, PresentSince: t1, AccountCreatedAt: ptrTime(t1)},
	}
	assert.Nil(t, userIDsToEnrichMissingAccounts(list, 10))
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
