/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ListChannelChattersRequest = {
    /**
     * Twitch OAuth account id (must exist; used for consistency with other channel APIs)
     */
    account_id: number;
    /**
     * Channel login
     */
    login: string;
    /**
     * When set, message_count is the number of persisted chat lines per chatter since this time (e.g. current stream started_at)
     */
    session_started_at?: string | null;
};

