/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type CreateTwitchAccountRequest = {
    /**
     * Twitch user id for this OAuth identity (same as Helix user id)
     */
    id: number;
    username: string;
    refresh_token: string;
    account_type?: CreateTwitchAccountRequest.account_type;
};
export namespace CreateTwitchAccountRequest {
    export enum account_type {
        MAIN = 'main',
        BOT = 'bot',
    }
}

