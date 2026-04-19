DROP TABLE IF EXISTS rules;

CREATE TABLE rules (
    id BIGSERIAL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT true,
    event_type TEXT NOT NULL,
    event_settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    middlewares JSONB NOT NULL DEFAULT '[]'::jsonb,
    action_type TEXT NOT NULL,
    action_settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    use_shared_pool BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_enabled_event ON rules (enabled, event_type);
