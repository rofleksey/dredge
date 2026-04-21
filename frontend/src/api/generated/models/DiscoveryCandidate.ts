/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { TwitchUser } from './TwitchUser';
export type DiscoveryCandidate = {
    user: TwitchUser;
    discovered_at: string;
    last_seen_at: string;
    viewer_count?: number | null;
    title?: string | null;
    game_name?: string | null;
    /**
     * Stream tag snapshot from the last matching Helix /streams row
     */
    stream_tags: Array<string>;
};

