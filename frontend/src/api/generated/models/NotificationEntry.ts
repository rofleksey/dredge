/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type NotificationEntry = {
    id: number;
    provider: NotificationEntry.provider;
    settings: Record<string, any>;
    enabled: boolean;
    created_at: string;
};
export namespace NotificationEntry {
    export enum provider {
        TELEGRAM = 'telegram',
        WEBHOOK = 'webhook',
    }
}

