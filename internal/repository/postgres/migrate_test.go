package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestFirstExecError(t *testing.T) {
	t.Parallel()

	require.NoError(t, firstExecError(nil))
	require.NoError(t, firstExecError([]*pgconn.Result{}))

	require.NoError(t, firstExecError([]*pgconn.Result{{}}))

	want := errors.New("exec failed")
	require.Equal(t, want, firstExecError([]*pgconn.Result{{Err: want}}))
	require.Equal(t, want, firstExecError([]*pgconn.Result{{}, {Err: want}}))
}
