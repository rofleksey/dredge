/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * notify — deliver `action_settings.text` via notification providers.
 * send_chat — post to the event channel via Helix; `action_settings` requires `message` (template).
 * OAuth account is chosen server-side (bot account if linked, else first linked account).
 *
 */
export enum RuleActionType {
    NOTIFY = 'notify',
    SEND_CHAT = 'send_chat',
}
