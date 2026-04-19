package handler

import "testing"

func TestToolResultPublicContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: ""},
		{name: "success object", raw: `{"rules":[]}`, want: ""},
		{name: "error string", raw: `{"error":"nope"}`, want: "nope"},
		{name: "error number", raw: `{"error":42}`, want: "42"},
		{name: "non-json", raw: `not json`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toolResultPublicContent(tt.raw)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
