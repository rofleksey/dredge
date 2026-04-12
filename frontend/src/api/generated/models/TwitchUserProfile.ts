/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { FollowedChannelEntry } from './FollowedChannelEntry';
import type { FollowedMonitoredChannel } from './FollowedMonitoredChannel';
export type TwitchUserProfile = {
    /**
     * Twitch numeric user id
     */
    id: number;
    username: string;
    monitored: boolean;
    marked: boolean;
    is_sus: boolean;
    sus_type?: string | null;
    sus_description?: string | null;
    sus_auto_suppressed: boolean;
    /**
     * Number of persisted chat messages attributed to this user as chatter
     */
    message_count: number;
    /**
     * Sum of IRC chat presence intervals since Monday 00:00 UTC this week
     */
    presence_seconds_this_week: number;
    /**
     * Twitch account creation time from Helix (when populated by enrichment)
     */
    account_created_at?: string | null;
    followed_monitored_channels?: Array<FollowedMonitoredChannel>;
    /**
     * Full list of channels this user follows (from GQL enrichment)
     */
    followed_channels: Array<FollowedChannelEntry>;
    /**
     * Current global blacklist (for client-side pinning and filtering)
     */
    channel_blacklist: Array<string>;
};

