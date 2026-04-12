/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type GetTwitchUserActivityTimelineRequest = {
    id: number;
    /**
     * Range start (RFC3339). Defaults to 7 days before `to`.
     */
    from?: string;
    /**
     * Range end (RFC3339). Defaults to now.
     */
    to?: string;
};

