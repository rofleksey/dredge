package ai

import (
	"testing"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestMergeRulePatch_partialDoesNotWipe(t *testing.T) {
	existing := entity.Rule{
		ID:        1,
		Name:      "n",
		Enabled:   true,
		EventType: "chat_message",
		EventSettings: map[string]any{
			"foo": "bar",
		},
		ActionType: "notify",
		ActionSettings: map[string]any{
			"text": "hello",
		},
		UseSharedPool: true,
	}
	raw := map[string]any{
		"id":      float64(1),
		"enabled": false,
	}
	got := mergeRulePatch(existing, raw)
	if got.Name != "n" {
		t.Fatalf("name: got %q", got.Name)
	}
	if got.Enabled {
		t.Fatal("enabled should be false")
	}
	if got.EventType != "chat_message" {
		t.Fatalf("event_type: got %q", got.EventType)
	}
	if got.EventSettings["foo"] != "bar" {
		t.Fatal("event_settings should be preserved")
	}
	if got.ActionType != "notify" {
		t.Fatalf("action_type: got %q", got.ActionType)
	}
}

func TestMergeRulePatch_eventSettingsMerge(t *testing.T) {
	existing := entity.Rule{
		EventType: "interval",
		EventSettings: map[string]any{
			"interval_seconds": float64(60),
			"channel":          "old",
		},
		ActionType:     "notify",
		ActionSettings: map[string]any{},
	}
	raw := map[string]any{
		"event_settings": map[string]any{
			"interval_seconds": float64(120),
		},
	}
	got := mergeRulePatch(existing, raw)
	if got.EventSettings["channel"] != "old" {
		t.Fatal("channel should remain when merging event_settings")
	}
	if got.EventSettings["interval_seconds"] != float64(120) {
		t.Fatalf("interval_seconds: %#v", got.EventSettings["interval_seconds"])
	}
}
