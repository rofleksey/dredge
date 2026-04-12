/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type RecordedStream = {
    id: number;
    /**
     * twitch_users.id of the broadcaster
     */
    channel_id: number;
    channel_login: string;
    helix_stream_id: string;
    started_at: string;
    /**
     * Null while the broadcast is still live
     */
    ended_at?: string | null;
    title?: string | null;
    game_name?: string | null;
    created_at: string;
};

