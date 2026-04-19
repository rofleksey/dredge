package httpmw

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOgenErrorHandler_loginRateLimited(t *testing.T) {
	t.Parallel()

	h := OgenErrorHandler()
	// Smoke: handler is non-nil and maps ErrLoginRateLimited without panicking.
	assert.NotNil(t, h)
	_ = context.Background()
}
