/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type AiMessage = {
    id: number;
    conversation_id: number;
    role: AiMessage.role;
    content: string;
    metadata: Record<string, any>;
    created_at: string;
};
export namespace AiMessage {
    export enum role {
        USER = 'user',
        ASSISTANT = 'assistant',
        TOOL = 'tool',
    }
}

