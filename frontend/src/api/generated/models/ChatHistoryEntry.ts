/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type ChatHistoryEntry = {
    id: number;
    /**
     * Channel login where the message appeared
     */
    channel: string;
    user: string;
    /**
     * Twitch user id of the chatter when known
     */
    chatter_user_id?: number | null;
    /**
     * Whether the chatter is marked in dredge
     */
    chatter_marked: boolean;
    /**
     * Whether the chatter is flagged as suspicious
     */
    chatter_is_sus: boolean;
    message: string;
    keyword_match: boolean;
    /**
     * irc: observed in chat; sent: posted via dredge
     */
    source: ChatHistoryEntry.source;
    created_at: string;
    /**
     * Twitch chat roles / badges for display (e.g. mod, VIP, verified bot, other badges)
     */
    badge_tags: Array<'moderator' | 'vip' | 'bot' | 'other'>;
};
export namespace ChatHistoryEntry {
    /**
     * irc: observed in chat; sent: posted via dredge
     */
    export enum source {
        IRC = 'irc',
        SENT = 'sent',
    }
}

