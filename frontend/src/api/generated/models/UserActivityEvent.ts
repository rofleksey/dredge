/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type UserActivityEvent = {
    id: number;
    /**
     * Chatter login (profile user)
     */
    username: string;
    event_type: UserActivityEvent.event_type;
    /**
     * Channel login when event is tied to a channel
     */
    channel?: string | null;
    details?: Record<string, any> | null;
    created_at: string;
};
export namespace UserActivityEvent {
    export enum event_type {
        CHAT_ONLINE = 'chat_online',
        CHAT_OFFLINE = 'chat_offline',
        MESSAGE = 'message',
    }
}

