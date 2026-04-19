package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_TestRuleRegex_match(t *testing.T) {
	t.Parallel()

	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	req := &gen.TestRuleRegexRequest{}
	req.SetPattern(`hello`)
	req.SetSample("say hello world")

	res, err := h.TestRuleRegex(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Matches)
	assert.True(t, res.CompileError.IsNull())
}

func TestHandler_TestRuleRegex_noMatch(t *testing.T) {
	t.Parallel()

	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	req := &gen.TestRuleRegexRequest{}
	req.SetPattern(`^foo$`)
	req.SetSample("bar")

	res, err := h.TestRuleRegex(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, res.Matches)
}

func TestHandler_TestRuleRegex_compileError(t *testing.T) {
	t.Parallel()

	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	req := &gen.TestRuleRegexRequest{}
	req.SetPattern(`(`)
	req.SetSample("x")

	res, err := h.TestRuleRegex(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, res.Matches)
	msg, ok := res.CompileError.Get()
	require.True(t, ok)
	assert.NotEmpty(t, msg)
}
