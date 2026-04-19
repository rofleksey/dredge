/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { RuleActionType } from './RuleActionType';
import type { RuleEventType } from './RuleEventType';
import type { RuleMiddleware } from './RuleMiddleware';
export type Rule = {
    id: number;
    enabled: boolean;
    event_type: RuleEventType;
    event_settings: Record<string, any>;
    middlewares: Array<RuleMiddleware>;
    action_type: RuleActionType;
    action_settings: Record<string, any>;
    use_shared_pool: boolean;
    created_at: string;
    updated_at: string;
};

