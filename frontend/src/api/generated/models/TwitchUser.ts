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
};

