package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_ListRuleTriggers(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	cur := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	curID := int64(123)
	rid := int64(42)

	repo.EXPECT().ListRuleTriggerEvents(gomock.Any(), entity.RuleTriggerListFilter{
		Limit:           20,
		CursorCreatedAt: &cur,
		CursorID:        &curID,
	}).Return([]entity.RuleTriggerEvent{
		{
			ID:           1,
			CreatedAt:    cur,
			RuleID:       &rid,
			RuleName:     "r1",
			TriggerEvent: "chat_message",
			ActionType:   "notify",
			DisplayText:  "[c] u: hi",
		},
	}, nil)

	out, err := h.ListRuleTriggers(adminCtx(), gen.ListRuleTriggersParams{
		Limit:           gen.NewOptInt(20),
		CursorCreatedAt: gen.NewOptDateTime(cur),
		CursorID:        gen.NewOptInt64(curID),
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, int64(1), out[0].ID)
	v, ok := out[0].RuleID.Get()
	assert.True(t, ok)
	assert.Equal(t, int64(42), v)
	assert.Equal(t, "r1", out[0].RuleName)
	assert.Equal(t, "[c] u: hi", out[0].DisplayText)
}
