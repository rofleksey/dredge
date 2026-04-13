/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type TwitchUser = {
    /**
     * Twitch numeric user id
     */
    id: number;
    /**
     * Canonical Twitch login (lowercase)
     */
    username: string;
    /**
     * When true, this channel is joined for IRC monitoring and keyword alerts
     */
    monitored: boolean;
    marked: boolean;
    is_sus: boolean;
    sus_type?: string | null;
    sus_description?: string | null;
    /**
     * When true, automatic suspicion will not mark this user until cleared in settings
     */
    sus_auto_suppressed: boolean;
    /**
     * When true, join IRC only while the channel has an active Helix stream
     */
    irc_only_when_live: boolean;
    /**
     * When irc_only_when_live is false, enable notifications for chat while the channel is offline on Twitch
     */
    notify_off_stream_messages: boolean;
    /**
     * When true, send notifications when this channel goes live on Twitch
     */
    notify_stream_start: boolean;
};

