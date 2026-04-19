package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestGenClientNoticeHelpers(t *testing.T) {
	t.Parallel()

	e := genClientNoticeErr("code", "msg")
	assert.Equal(t, gen.ClientNoticeSeverityError, e.Severity)
	assert.Equal(t, "code", e.Code)
	assert.Equal(t, "msg", e.Message)

	w := genClientNoticeWarn("w", "m")
	assert.Equal(t, gen.ClientNoticeSeverityWarning, w.Severity)
}
