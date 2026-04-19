package httptransport

import "testing"

func TestHandler_UpdateTwitchAccount_smoke(t *testing.T) {
	t.Parallel()
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()
	_ = h
}
