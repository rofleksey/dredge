/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ChannelChatterEntry = {
    /**
     * Chatter login (lowercase)
     */
    login: string;
    user_twitch_id: number;
    /**
     * When this chatter was first seen in the channel for the current stretch (IRC JOIN or NAMES sync)
     */
    present_since: string;
    /**
     * Twitch account creation from Helix when enrichment has populated it
     */
    account_created_at?: string | null;
    /**
     * Populated when session_started_at was sent in the request
     */
    message_count?: number | null;
};

