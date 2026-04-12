/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type CreateNotificationRequest = {
    provider: CreateNotificationRequest.provider;
    settings: Record<string, any>;
    enabled?: boolean;
};
export namespace CreateNotificationRequest {
    export enum provider {
        TELEGRAM = 'telegram',
        WEBHOOK = 'webhook',
    }
}

