/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type AiSettings = {
    /**
     * OpenAI-compatible API base URL (e.g. https://api.openai.com/v1)
     */
    base_url: string;
    model: string;
    /**
     * Whether an API token is stored (value is never returned)
     */
    has_token: boolean;
    /**
     * Last four characters of the stored token when present
     */
    token_last4?: string;
    updated_at: string;
};

