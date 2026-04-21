CREATE TABLE IF NOT EXISTS channel_discovery_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT false,
    poll_interval_seconds INT NOT NULL DEFAULT 3600 CHECK (poll_interval_seconds >= 60),
    game_id TEXT NOT NULL DEFAULT '',
    min_live_viewers INT NOT NULL DEFAULT 0 CHECK (min_live_viewers >= 0),
    required_stream_tags TEXT[] NOT NULL DEFAULT '{}',
    max_stream_pages_per_run INT NOT NULL DEFAULT 20 CHECK (max_stream_pages_per_run >= 1)
);

INSERT INTO channel_discovery_settings (id) VALUES (1)
    ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS twitch_discovery_candidates (
    twitch_user_id BIGINT PRIMARY KEY REFERENCES twitch_users (id) ON DELETE CASCADE,
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    viewer_count BIGINT,
    title TEXT,
    game_name TEXT,
    stream_tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_twitch_discovery_candidates_last_seen ON twitch_discovery_candidates (last_seen_at DESC);

CREATE TABLE IF NOT EXISTS twitch_discovery_denied (
    twitch_user_id BIGINT PRIMARY KEY REFERENCES twitch_users (id) ON DELETE CASCADE,
    denied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
