package handler

import "testing"

func TestHandler_GetSuspicionSettings_smoke(t *testing.T) {
	t.Parallel()
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()
	_ = h
}
