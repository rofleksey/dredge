CREATE TABLE IF NOT EXISTS messages
(
    id       VARCHAR(255) PRIMARY KEY,
    created  TIMESTAMP    NOT NULL,
    username VARCHAR(255) NOT NULL,
    channel  VARCHAR(255) NOT NULL,
    text     TEXT         NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel);
CREATE INDEX IF NOT EXISTS idx_messages_channel_username ON messages(channel, username);
CREATE INDEX IF NOT EXISTS idx_messages_channel_username_created ON messages(channel, username, created DESC);
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created DESC);
CREATE INDEX IF NOT EXISTS idx_messages_channel_created ON messages(channel, created DESC);
CREATE INDEX IF NOT EXISTS idx_messages_username_created ON messages(username, created DESC);

CREATE TABLE IF NOT EXISTS migration
(
    id      VARCHAR(255) PRIMARY KEY,
    applied TIMESTAMP NOT NULL
);
