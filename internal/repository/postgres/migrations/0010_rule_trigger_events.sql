CREATE TABLE IF NOT EXISTS rule_trigger_events (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    rule_id BIGINT REFERENCES rules (id) ON DELETE SET NULL,
    rule_name TEXT NOT NULL DEFAULT '',
    trigger_event TEXT NOT NULL,
    action_type TEXT NOT NULL,
    display_text TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS rule_trigger_events_created_at_id_idx
    ON rule_trigger_events (created_at DESC, id DESC);
