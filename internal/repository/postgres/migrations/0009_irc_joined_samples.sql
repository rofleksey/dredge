CREATE TABLE IF NOT EXISTS irc_joined_samples (
    id BIGSERIAL PRIMARY KEY,
    joined_count INT NOT NULL CHECK (joined_count >= 0),
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_irc_joined_samples_captured_at ON irc_joined_samples (captured_at);
