/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ChannelLive = {
    broadcaster_id: number;
    broadcaster_login: string;
    display_name: string;
    profile_image_url: string;
    is_live: boolean;
    title?: string | null;
    game_name?: string | null;
    viewer_count?: number | null;
    /**
     * Count of users in the IRC-maintained chat snapshot (channel_chatters). Compare to viewer_count for the delta shown in the watch UI.
     *
     */
    channel_chatter_count?: number | null;
    started_at?: string | null;
};

