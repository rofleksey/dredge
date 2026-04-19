package httptransport

import "testing"

func TestHandler_ListChannelBlacklist_smoke(t *testing.T) {
	t.Parallel()
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()
	_ = h
}
