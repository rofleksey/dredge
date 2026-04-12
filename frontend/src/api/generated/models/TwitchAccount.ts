/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type TwitchAccount = {
    id: number;
    username: string;
    account_type: TwitchAccount.account_type;
    created_at: string;
};
export namespace TwitchAccount {
    export enum account_type {
        MAIN = 'main',
        BOT = 'bot',
    }
}

