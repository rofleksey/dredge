/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type IrcMonitorStatus = {
    /**
     * IRC monitor TCP session is connected
     */
    connected: boolean;
    channels: Array<{
        login: string;
        /**
         * True when the IRC client's NAMES-backed userlist for this channel is non-empty (session state)
         */
        irc_ok: boolean;
    }>;
};

