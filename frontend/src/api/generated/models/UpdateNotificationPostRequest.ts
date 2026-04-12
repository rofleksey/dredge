/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type UpdateNotificationPostRequest = {
    id: number;
    provider?: UpdateNotificationPostRequest.provider;
    settings?: Record<string, any>;
    enabled?: boolean;
};
export namespace UpdateNotificationPostRequest {
    export enum provider {
        TELEGRAM = 'telegram',
        WEBHOOK = 'webhook',
    }
}

