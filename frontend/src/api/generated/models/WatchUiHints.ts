/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type WatchUiHints = {
    /**
     * How often the watch UI should refresh Helix stream metadata (viewer count, live state)
     */
    viewer_poll_interval_seconds: number;
    /**
     * How often the server refreshes IRC NAMES into channel_chatters
     */
    channel_chatters_sync_interval_seconds: number;
};

