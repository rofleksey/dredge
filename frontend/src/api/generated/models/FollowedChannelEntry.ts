/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type FollowedChannelEntry = {
    channel_id: number;
    channel_login: string;
    followed_at?: string | null;
    /**
     * True when this followed channel login is on the global blacklist
     */
    on_blacklist: boolean;
};

