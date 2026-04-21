/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ChannelDiscoverySettings = {
    /**
     * When true, the backend periodically scans Helix live streams for the configured game
     */
    enabled: boolean;
    /**
     * Minimum seconds between discovery runs (default 3600)
     */
    poll_interval_seconds: number;
    /**
     * Twitch Helix category id (game_id) passed to GET /helix/streams; required when enabled is true
     */
    game_id: string;
    /**
     * Minimum concurrent viewers (Helix stream viewer_count) for a channel to be suggested
     */
    min_live_viewers: number;
    /**
     * When non-empty, a live stream must include every listed tag on its Helix `tags` array (case-insensitive match).
     * Empty means no tag filter.
     *
     */
    required_stream_tags: Array<string>;
    /**
     * Max Helix /streams pages (100 streams each) per discovery run
     */
    max_stream_pages_per_run: number;
};

