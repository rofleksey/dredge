package handler

import "testing"

func TestHandler_GetIrcMonitorSettings_smoke(t *testing.T) {
	t.Parallel()
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()
	_ = h
}
