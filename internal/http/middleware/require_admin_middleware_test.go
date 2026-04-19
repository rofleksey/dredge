package httpmw

import "testing"

func TestRequireAdminMiddleware_returnsMiddleware(t *testing.T) {
	t.Parallel()
	m := RequireAdminMiddleware()
	if m == nil {
		t.Fatal("expected non-nil middleware")
	}
}
