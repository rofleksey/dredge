/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type WatchUiHints = {
    /**
     * How often the watch UI may poll for a manually entered (non-directory) channel via getChannelLive
     */
    viewer_poll_interval_seconds: number;
    /**
     * How often the server refreshes IRC NAMES into channel_chatters
     */
    channel_chatters_sync_interval_seconds: number;
    /**
     * How often the backend polls Helix (batched /streams) for monitored channels; use this interval when refreshing GET /twitch/users?monitored_only=true
     */
    monitored_live_poll_interval_seconds: number;
};

