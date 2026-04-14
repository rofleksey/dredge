ALTER TABLE irc_monitor_settings
    ADD COLUMN IF NOT EXISTS enrichment_cooldown_seconds BIGINT NOT NULL DEFAULT 86400;

UPDATE irc_monitor_settings
SET enrichment_cooldown_seconds = 86400
WHERE enrichment_cooldown_seconds <= 0;
