ALTER TABLE twitch_users
    ALTER COLUMN irc_only_when_live SET DEFAULT true,
    ALTER COLUMN notify_off_stream_messages SET DEFAULT false,
    ALTER COLUMN notify_stream_start SET DEFAULT false;

UPDATE twitch_users
SET irc_only_when_live = true,
    notify_off_stream_messages = false,
    notify_stream_start = false;
