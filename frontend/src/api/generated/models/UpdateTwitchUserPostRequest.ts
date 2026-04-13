/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type UpdateTwitchUserPostRequest = {
    id: number;
    monitored?: boolean;
    marked?: boolean;
    is_sus?: boolean;
    sus_type?: string | null;
    sus_description?: string | null;
    /**
     * Set true when clearing suspicion to block automatic re-marking; set false to allow auto detection again
     */
    sus_auto_suppressed?: boolean;
    irc_only_when_live?: boolean;
    notify_off_stream_messages?: boolean;
    notify_stream_start?: boolean;
};

