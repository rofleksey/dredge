/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type SuspicionSettings = {
    auto_check_account_age: boolean;
    /**
     * Accounts newer than this many days are suspicious
     */
    account_age_sus_days: number;
    auto_check_blacklist: boolean;
    auto_check_low_follows: boolean;
    /**
     * Suspicious when total follow count is strictly below this value
     */
    low_follows_threshold: number;
    /**
     * Safety cap when paginating GQL follows per user (default 1 page)
     */
    max_gql_follow_pages: number;
};

