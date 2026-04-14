CREATE TABLE IF NOT EXISTS twitch_accounts (
    id BIGINT PRIMARY KEY,
    username TEXT NOT NULL CHECK (username = lower(username)),
    refresh_token TEXT NOT NULL,
    account_type TEXT NOT NULL DEFAULT 'main' CHECK (account_type IN ('main', 'bot')),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_twitch_accounts_username_active ON twitch_accounts (username) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS twitch_users (
    id BIGINT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL CHECK (username = lower(username)),
    monitored BOOLEAN NOT NULL DEFAULT true,
    marked BOOLEAN NOT NULL DEFAULT false,
    is_sus BOOLEAN NOT NULL DEFAULT false,
    sus_type TEXT,
    sus_description TEXT,
    sus_auto_suppressed BOOLEAN NOT NULL DEFAULT false,
    irc_only_when_live BOOLEAN NOT NULL DEFAULT true,
    notify_off_stream_messages BOOLEAN NOT NULL DEFAULT false,
    notify_stream_start BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS rules (
    id BIGSERIAL PRIMARY KEY,
    regex TEXT NOT NULL,
    included_users TEXT NOT NULL DEFAULT '*',
    denied_users TEXT NOT NULL DEFAULT '',
    included_channels TEXT NOT NULL DEFAULT '*',
    denied_channels TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS notification_entries (
    id BIGSERIAL PRIMARY KEY,
    provider TEXT NOT NULL CHECK (provider IN ('telegram', 'webhook')),
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS streams (
    id BIGSERIAL PRIMARY KEY,
    channel_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    helix_stream_id TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    title TEXT,
    game_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT streams_helix_stream_id_unique UNIQUE (helix_stream_id)
);

CREATE INDEX IF NOT EXISTS idx_streams_channel_started ON streams (channel_twitch_user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_streams_channel_open ON streams (channel_twitch_user_id) WHERE ended_at IS NULL;

CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE RESTRICT,
    chatter_twitch_user_id BIGINT REFERENCES twitch_users (id) ON DELETE SET NULL,
    username TEXT NOT NULL CHECK (username = lower(username)),
    body TEXT NOT NULL,
    keyword_match BOOLEAN NOT NULL DEFAULT false,
    first_message BOOLEAN NOT NULL DEFAULT false,
    msg_type TEXT NOT NULL DEFAULT 'irc',
    badge_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stream_id BIGINT REFERENCES streams (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_twitch_user_created ON chat_messages (twitch_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_id_desc ON chat_messages (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_chatter_created ON chat_messages (chatter_twitch_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_stream ON chat_messages (stream_id) WHERE stream_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS channel_chatters (
    channel_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    chatter_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    present_since TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_twitch_user_id, chatter_twitch_user_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_chatters_channel ON channel_chatters (channel_twitch_user_id);

CREATE TABLE IF NOT EXISTS user_activity_events (
    id BIGSERIAL PRIMARY KEY,
    chatter_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    channel_twitch_user_id BIGINT REFERENCES twitch_users (id) ON DELETE SET NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_activity_events_chatter_created ON user_activity_events (chatter_twitch_user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_user_activity_events_channel_created ON user_activity_events (channel_twitch_user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS twitch_user_helix_meta (
    twitch_user_id BIGINT PRIMARY KEY REFERENCES twitch_users (id) ON DELETE CASCADE,
    account_created_at TIMESTAMPTZ,
    helix_fetched_at TIMESTAMPTZ,
    profile_image_url TEXT
);

CREATE TABLE IF NOT EXISTS twitch_user_channel_follows (
    chatter_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    channel_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    followed_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chatter_twitch_user_id, channel_twitch_user_id)
);

CREATE TABLE IF NOT EXISTS user_followed_channels (
    follower_twitch_user_id BIGINT NOT NULL REFERENCES twitch_users (id) ON DELETE CASCADE,
    followed_channel_id BIGINT NOT NULL,
    followed_channel_login TEXT NOT NULL CHECK (followed_channel_login = lower(followed_channel_login)),
    followed_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_twitch_user_id, followed_channel_id)
);

CREATE INDEX IF NOT EXISTS idx_user_followed_channels_follower ON user_followed_channels (follower_twitch_user_id);

CREATE TABLE IF NOT EXISTS channel_blacklist (
    login TEXT PRIMARY KEY CHECK (login = lower(login))
);

CREATE TABLE IF NOT EXISTS suspicion_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    auto_check_account_age BOOLEAN NOT NULL DEFAULT true,
    account_age_sus_days INT NOT NULL DEFAULT 14,
    auto_check_blacklist BOOLEAN NOT NULL DEFAULT true,
    auto_check_low_follows BOOLEAN NOT NULL DEFAULT true,
    low_follows_threshold INT NOT NULL DEFAULT 10,
    max_gql_follow_pages INT NOT NULL DEFAULT 1
);

INSERT INTO suspicion_settings (id) VALUES (1)
    ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS irc_monitor_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    oauth_twitch_account_id BIGINT REFERENCES twitch_accounts (id) ON DELETE SET NULL,
    enrichment_cooldown_seconds BIGINT NOT NULL DEFAULT 86400
);

INSERT INTO irc_monitor_settings (id, oauth_twitch_account_id, enrichment_cooldown_seconds) VALUES (1, NULL, 86400)
    ON CONFLICT (id) DO NOTHING;