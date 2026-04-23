/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type SystemStatsTables = {
    twitch_users: number;
    /**
     * twitch_accounts rows with deleted_at IS NULL
     */
    twitch_accounts_active: number;
    twitch_accounts_all: number;
    rules: number;
    notification_entries: number;
    streams: number;
    /**
     * streams with ended_at IS NULL
     */
    streams_open: number;
    chat_messages: number;
    channel_chatters: number;
    user_activity_events: number;
    twitch_user_helix_meta: number;
    twitch_user_channel_follows: number;
    user_followed_channels: number;
    channel_blacklist: number;
    rule_trigger_events: number;
    irc_joined_samples: number;
    twitch_discovery_candidates: number;
    twitch_discovery_denied: number;
    ai_conversations: number;
    ai_messages: number;
};

