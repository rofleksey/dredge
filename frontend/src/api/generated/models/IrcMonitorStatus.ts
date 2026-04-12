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
         * True after the monitor has received self-JOIN for this channel
         */
        irc_ok: boolean;
    }>;
};

