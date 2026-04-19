package handler

import "testing"

func TestHandler_ListRecordedStreamActivity_smoke(t *testing.T) {
	t.Parallel()
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()
	_ = h
}
