package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrNoSentry(t *testing.T) {
	t.Parallel()

	assert.True(t, errors.Is(ErrNoSentry, ErrNoSentry))
}
