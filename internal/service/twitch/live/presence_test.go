package live

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsGoTwitchIRCUserlistMissing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "other", err: errors.New("network down"), want: false},
		{
			name: "library_message",
			err:  fmt.Errorf("Could not find userlist for channel 'foo' in client"),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := isGoTwitchIRCUserlistMissing(tc.err); got != tc.want {
				t.Fatalf("isGoTwitchIRCUserlistMissing(...) = %v, want %v", got, tc.want)
			}
		})
	}
}
