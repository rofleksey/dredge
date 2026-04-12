package postgres

import (
	"testing"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors_distinct(t *testing.T) {
	t.Parallel()

	errs := []error{
		entity.ErrRuleNotFound,
		entity.ErrNotificationNotFound,
		entity.ErrTwitchAccountNotFound,
		entity.ErrTwitchUserNotFound,
		entity.ErrNoTwitchUserForChannel,
	}

	for i := range errs {
		for j := i + 1; j < len(errs); j++ {
			assert.NotErrorIs(t, errs[i], errs[j], "errors at %d and %d must differ", i, j)
		}
	}
}
