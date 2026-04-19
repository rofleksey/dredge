package helix

import (
	"encoding/json"
	"fmt"
)

func helixAPIFailure(status int, body []byte) error {
	var wrap struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &wrap); err == nil && wrap.Message != "" {
		return &ChatSendError{Message: wrap.Message}
	}
	return fmt.Errorf("helix: status %d: %s", status, string(body))
}
