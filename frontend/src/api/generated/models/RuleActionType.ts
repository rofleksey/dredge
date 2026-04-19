/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * notify — deliver `action_settings.text` via notification providers.
 * send_chat — post to the event channel via Helix; `action_settings` requires `message` (template).
 * Optional `account_id` (integer, app-linked Twitch OAuth row id) selects which linked account sends the message;
 * if omitted or zero, the server uses the linked bot account when present, otherwise the first linked account.
 *
 */
export enum RuleActionType {
    NOTIFY = 'notify',
    SEND_CHAT = 'send_chat',
}
