/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type UpdateNotificationRequest = {
    provider?: UpdateNotificationRequest.provider;
    settings?: Record<string, any>;
    enabled?: boolean;
};
export namespace UpdateNotificationRequest {
    export enum provider {
        TELEGRAM = 'telegram',
        WEBHOOK = 'webhook',
    }
}

