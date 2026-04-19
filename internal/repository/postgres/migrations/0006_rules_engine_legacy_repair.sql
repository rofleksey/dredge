-- Replace legacy `rules` (from 0001_init) when 0005 did not run or left an old shape.
DO $body$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'rules'
          AND column_name = 'enabled'
    ) THEN
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
    END IF;
END;
$body$;
