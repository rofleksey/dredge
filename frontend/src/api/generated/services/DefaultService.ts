/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Account } from '../models/Account';
import type { ActivityTimelineSegment } from '../models/ActivityTimelineSegment';
import type { AiConversation } from '../models/AiConversation';
import type { AiMessage } from '../models/AiMessage';
import type { AiRunAccepted } from '../models/AiRunAccepted';
import type { AiSettings } from '../models/AiSettings';
import type { ChannelBlacklistChange } from '../models/ChannelBlacklistChange';
import type { ChannelChatterEntry } from '../models/ChannelChatterEntry';
import type { ChannelLive } from '../models/ChannelLive';
import type { ChatHistoryEntry } from '../models/ChatHistoryEntry';
import type { ConfirmAiToolRequest } from '../models/ConfirmAiToolRequest';
import type { CountResponse } from '../models/CountResponse';
import type { CreateAiConversationRequest } from '../models/CreateAiConversationRequest';
import type { CreateAiMessageRequest } from '../models/CreateAiMessageRequest';
import type { CreateNotificationRequest } from '../models/CreateNotificationRequest';
import type { CreateRuleRequest } from '../models/CreateRuleRequest';
import type { CreateTwitchAccountRequest } from '../models/CreateTwitchAccountRequest';
import type { CreateTwitchUserRequest } from '../models/CreateTwitchUserRequest';
import type { DeleteByIDRequest } from '../models/DeleteByIDRequest';
import type { GetChannelLiveRequest } from '../models/GetChannelLiveRequest';
import type { GetTwitchUserActivityTimelineRequest } from '../models/GetTwitchUserActivityTimelineRequest';
import type { GetTwitchUserProfileRequest } from '../models/GetTwitchUserProfileRequest';
import type { IrcMonitorSettings } from '../models/IrcMonitorSettings';
import type { IrcMonitorStatus } from '../models/IrcMonitorStatus';
import type { ListChannelChattersRequest } from '../models/ListChannelChattersRequest';
import type { ListTwitchUserActivityRequest } from '../models/ListTwitchUserActivityRequest';
import type { LoginRequest } from '../models/LoginRequest';
import type { LoginResponse } from '../models/LoginResponse';
import type { NotificationEntry } from '../models/NotificationEntry';
import type { PatchAiSettingsRequest } from '../models/PatchAiSettingsRequest';
import type { RecordedStream } from '../models/RecordedStream';
import type { Rule } from '../models/Rule';
import type { RuleTemplateVariablesResponse } from '../models/RuleTemplateVariablesResponse';
import type { SendMessageRequest } from '../models/SendMessageRequest';
import type { StartTwitchOAuthRequest } from '../models/StartTwitchOAuthRequest';
import type { StartTwitchOAuthResponse } from '../models/StartTwitchOAuthResponse';
import type { StreamLeaderboardEntry } from '../models/StreamLeaderboardEntry';
import type { StreamLeaderboardSort } from '../models/StreamLeaderboardSort';
import type { SuspicionSettings } from '../models/SuspicionSettings';
import type { TestRuleRegexRequest } from '../models/TestRuleRegexRequest';
import type { TestRuleRegexResponse } from '../models/TestRuleRegexResponse';
import type { TwitchAccount } from '../models/TwitchAccount';
import type { TwitchUser } from '../models/TwitchUser';
import type { TwitchUserProfile } from '../models/TwitchUserProfile';
import type { UpdateNotificationPostRequest } from '../models/UpdateNotificationPostRequest';
import type { UpdateRulePostRequest } from '../models/UpdateRulePostRequest';
import type { UpdateTwitchAccountPostRequest } from '../models/UpdateTwitchAccountPostRequest';
import type { UpdateTwitchUserPostRequest } from '../models/UpdateTwitchUserPostRequest';
import type { UserActivityEvent } from '../models/UserActivityEvent';
import type { WatchUiHints } from '../models/WatchUiHints';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * @returns LoginResponse Login success
     * @throws ApiError
     */
    public static login({
        requestBody,
    }: {
        requestBody: LoginRequest,
    }): CancelablePromise<LoginResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/login',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                401: `Unauthorized`,
            },
        });
    }
    /**
     * @returns Account Current account
     * @throws ApiError
     */
    public static me(): CancelablePromise<Account> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/me',
            errors: {
                401: `Unauthorized`,
            },
        });
    }
    /**
     * @returns TwitchUser Twitch users (streamers/channels); monitored flag controls IRC monitoring
     * @throws ApiError
     */
    public static listTwitchUsers({
        monitoredOnly = false,
    }: {
        /**
         * When true, return only monitored channels (same TwitchUser fields as the full list; profile_image_url and channel_live are not populated on this path).
         */
        monitoredOnly?: boolean,
    }): CancelablePromise<Array<TwitchUser>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-users',
            query: {
                'monitored_only': monitoredOnly,
            },
        });
    }
    /**
     * @returns TwitchUser Created
     * @throws ApiError
     */
    public static createTwitchUser({
        requestBody,
    }: {
        requestBody: CreateTwitchUserRequest,
    }): CancelablePromise<TwitchUser> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-users',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request (e.g. unknown Twitch channel)`,
            },
        });
    }
    /**
     * @returns TwitchUser Updated
     * @throws ApiError
     */
    public static updateTwitchUser({
        requestBody,
    }: {
        requestBody: UpdateTwitchUserPostRequest,
    }): CancelablePromise<TwitchUser> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-users/update',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid monitor settings (e.g. off-stream notifications require disabling live-only IRC)`,
                404: `Twitch user not found`,
            },
        });
    }
    /**
     * @returns string Blacklisted channel logins (lowercase)
     * @throws ApiError
     */
    public static listChannelBlacklist(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/channel-blacklist',
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static setChannelBlacklist({
        requestBody,
    }: {
        requestBody: ChannelBlacklistChange,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/channel-blacklist',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
            },
        });
    }
    /**
     * @returns SuspicionSettings Suspicion thresholds
     * @throws ApiError
     */
    public static getSuspicionSettings(): CancelablePromise<SuspicionSettings> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/suspicion-settings',
        });
    }
    /**
     * @returns SuspicionSettings Updated settings
     * @throws ApiError
     */
    public static updateSuspicionSettings({
        requestBody,
    }: {
        requestBody: SuspicionSettings,
    }): CancelablePromise<SuspicionSettings> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/settings/suspicion-settings',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns IrcMonitorSettings IRC monitor identity (anonymous vs linked OAuth account)
     * @throws ApiError
     */
    public static getIrcMonitorSettings(): CancelablePromise<IrcMonitorSettings> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/irc-monitor-settings',
        });
    }
    /**
     * @returns IrcMonitorSettings Updated settings
     * @throws ApiError
     */
    public static updateIrcMonitorSettings({
        requestBody,
    }: {
        requestBody: IrcMonitorSettings,
    }): CancelablePromise<IrcMonitorSettings> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/settings/irc-monitor-settings',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns Rule Rule list
     * @throws ApiError
     */
    public static listRules(): CancelablePromise<Array<Rule>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/rules',
        });
    }
    /**
     * @returns Rule Created
     * @throws ApiError
     */
    public static createRule({
        requestBody,
    }: {
        requestBody: CreateRuleRequest,
    }): CancelablePromise<Rule> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/rules',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns CountResponse Total rule count
     * @throws ApiError
     */
    public static countRules(): CancelablePromise<CountResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/rules/count',
        });
    }
    /**
     * List rule message template placeholders
     * Names and descriptions for `$NAME` placeholders in notify and send_chat message templates.
     *
     * @returns RuleTemplateVariablesResponse Placeholder metadata
     * @throws ApiError
     */
    public static listRuleTemplateVariables(): CancelablePromise<RuleTemplateVariablesResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/rules/template-variables',
        });
    }
    /**
     * @returns Rule Updated
     * @throws ApiError
     */
    public static updateRule({
        requestBody,
    }: {
        requestBody: UpdateRulePostRequest,
    }): CancelablePromise<Rule> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/rules/update',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Rule not found`,
            },
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static deleteRule({
        requestBody,
    }: {
        requestBody: DeleteByIDRequest,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/rules/delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Rule not found`,
            },
        });
    }
    /**
     * @returns TestRuleRegexResponse Match result or compile error
     * @throws ApiError
     */
    public static testRuleRegex({
        requestBody,
    }: {
        requestBody: TestRuleRegexRequest,
    }): CancelablePromise<TestRuleRegexResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/rules/test-regex',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * List notification entries (newest first) with cursor-based incremental loading.
     * @returns NotificationEntry Notification entries
     * @throws ApiError
     */
    public static listNotifications({
        limit = 50,
        cursorCreatedAt,
        cursorId,
    }: {
        limit?: number,
        /**
         * Keyset cursor; use with cursor_id from the last entry of the previous batch.
         */
        cursorCreatedAt?: string,
        cursorId?: number,
    }): CancelablePromise<Array<NotificationEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/notifications',
            query: {
                'limit': limit,
                'cursor_created_at': cursorCreatedAt,
                'cursor_id': cursorId,
            },
        });
    }
    /**
     * @returns NotificationEntry Created
     * @throws ApiError
     */
    public static createNotification({
        requestBody,
    }: {
        requestBody: CreateNotificationRequest,
    }): CancelablePromise<NotificationEntry> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/notifications',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns NotificationEntry Updated
     * @throws ApiError
     */
    public static updateNotification({
        requestBody,
    }: {
        requestBody: UpdateNotificationPostRequest,
    }): CancelablePromise<NotificationEntry> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/notifications/update',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Not found`,
            },
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static deleteNotification({
        requestBody,
    }: {
        requestBody: DeleteByIDRequest,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/notifications/delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Not found`,
            },
        });
    }
    /**
     * @returns TwitchAccount Twitch accounts
     * @throws ApiError
     */
    public static listTwitchAccounts(): CancelablePromise<Array<TwitchAccount>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-accounts',
        });
    }
    /**
     * @returns TwitchAccount Created
     * @throws ApiError
     */
    public static createTwitchAccount({
        requestBody,
    }: {
        requestBody: CreateTwitchAccountRequest,
    }): CancelablePromise<TwitchAccount> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-accounts',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns CountResponse Active Twitch account count
     * @throws ApiError
     */
    public static countTwitchAccounts(): CancelablePromise<CountResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-accounts/count',
        });
    }
    /**
     * @returns TwitchAccount Updated
     * @throws ApiError
     */
    public static updateTwitchAccount({
        requestBody,
    }: {
        requestBody: UpdateTwitchAccountPostRequest,
    }): CancelablePromise<TwitchAccount> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-accounts/update',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Twitch account not found`,
            },
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static deleteTwitchAccount({
        requestBody,
    }: {
        requestBody: DeleteByIDRequest,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-accounts/delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Twitch account not found`,
            },
        });
    }
    /**
     * Start Twitch authorization (browser) to link an account without pasting a refresh token
     * @returns StartTwitchOAuthResponse Open this URL in a browser to authorize on Twitch
     * @throws ApiError
     */
    public static startTwitchOAuth({
        requestBody,
    }: {
        requestBody?: StartTwitchOAuthRequest,
    }): CancelablePromise<StartTwitchOAuthResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-accounts/oauth/start',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns any Accepted
     * @throws ApiError
     */
    public static sendMessage({
        requestBody,
    }: {
        requestBody: SendMessageRequest,
    }): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/send',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request`,
                401: `Unauthorized`,
                422: `Send failed (e.g. token or IRC)`,
                502: `Upstream or connection failure`,
            },
        });
    }
    /**
     * @returns ChatHistoryEntry Recent chat messages for the channel (oldest first)
     * @throws ApiError
     */
    public static listChatHistory({
        channel,
        limit = 50,
    }: {
        channel: string,
        limit?: number,
    }): CancelablePromise<Array<ChatHistoryEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/chat/history',
            query: {
                'channel': channel,
                'limit': limit,
            },
            errors: {
                404: `Channel is not monitored`,
            },
        });
    }
    /**
     * Search persisted chat messages (newest first). Omit filters to list recent messages.
     * @returns ChatHistoryEntry Matching messages
     * @throws ApiError
     */
    public static listTwitchMessages({
        limit = 50,
        cursorCreatedAt,
        cursorId,
        username,
        text,
        channel,
        createdFrom,
        createdTo,
        chatterUserId,
    }: {
        limit?: number,
        /**
         * Keyset cursor; use with cursor_id from the last message of the previous page
         */
        cursorCreatedAt?: string,
        cursorId?: number,
        /**
         * Filter by chatter login (substring match, case-insensitive)
         */
        username?: string,
        /**
         * Filter by message body (substring match, case-insensitive)
         */
        text?: string,
        /**
         * Filter by channel login
         */
        channel?: string,
        createdFrom?: string,
        createdTo?: string,
        /**
         * Filter by Twitch numeric user id of the chatter
         */
        chatterUserId?: number,
    }): CancelablePromise<Array<ChatHistoryEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/messages',
            query: {
                'limit': limit,
                'cursor_created_at': cursorCreatedAt,
                'cursor_id': cursorId,
                'username': username,
                'text': text,
                'channel': channel,
                'created_from': createdFrom,
                'created_to': createdTo,
                'chatter_user_id': chatterUserId,
            },
        });
    }
    /**
     * Count messages matching the same filters as list (ignores limit/cursor).
     * @returns CountResponse Count
     * @throws ApiError
     */
    public static countTwitchMessages({
        username,
        text,
        channel,
        createdFrom,
        createdTo,
        chatterUserId,
    }: {
        username?: string,
        text?: string,
        channel?: string,
        createdFrom?: string,
        createdTo?: string,
        chatterUserId?: number,
    }): CancelablePromise<CountResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/messages/count',
            query: {
                'username': username,
                'text': text,
                'channel': channel,
                'created_from': createdFrom,
                'created_to': createdTo,
                'chatter_user_id': chatterUserId,
            },
        });
    }
    /**
     * List known Twitch users (chatters and channels) for directory search
     * @returns TwitchUser Users (newest id first)
     * @throws ApiError
     */
    public static listTwitchDirectoryUsers({
        username,
        limit = 50,
        cursorId,
        monitoredOnly = false,
    }: {
        /**
         * Substring match on login (case-insensitive)
         */
        username?: string,
        limit?: number,
        /**
         * Keyset cursor (twitch user id from the last row of the previous page; sort is id desc)
         */
        cursorId?: number,
        /**
         * When true, only monitored channels are returned with channel_live filled from the server's batched Helix /streams poll.
         */
        monitoredOnly?: boolean,
    }): CancelablePromise<Array<TwitchUser>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/users',
            query: {
                'username': username,
                'limit': limit,
                'cursor_id': cursorId,
                'monitored_only': monitoredOnly,
            },
        });
    }
    /**
     * @returns CountResponse Count
     * @throws ApiError
     */
    public static countTwitchDirectoryUsers({
        username,
        monitoredOnly = false,
    }: {
        username?: string,
        /**
         * When true, count only monitored channels (same filter as list).
         */
        monitoredOnly?: boolean,
    }): CancelablePromise<CountResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/users/count',
            query: {
                'username': username,
                'monitored_only': monitoredOnly,
            },
        });
    }
    /**
     * @returns TwitchUserProfile User profile
     * @throws ApiError
     */
    public static getTwitchUserProfile({
        requestBody,
    }: {
        requestBody: GetTwitchUserProfileRequest,
    }): CancelablePromise<TwitchUserProfile> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/users/profile',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `User not found`,
            },
        });
    }
    /**
     * @returns ChannelLive Channel profile and live state
     * @throws ApiError
     */
    public static getChannelLive({
        requestBody,
    }: {
        requestBody: GetChannelLiveRequest,
    }): CancelablePromise<ChannelLive> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/channels/live',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Unknown channel`,
            },
        });
    }
    /**
     * @returns ChannelChatterEntry Chat participants from the IRC-maintained channel_chatters snapshot
     * @throws ApiError
     */
    public static listChannelChatters({
        requestBody,
    }: {
        requestBody: ListChannelChattersRequest,
    }): CancelablePromise<Array<ChannelChatterEntry>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/channels/chatters',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Twitch account or channel not found`,
            },
        });
    }
    /**
     * @returns WatchUiHints SPA polling intervals derived from server config
     * @throws ApiError
     */
    public static getWatchUiHints(): CancelablePromise<WatchUiHints> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/watch/hints',
        });
    }
    /**
     * @returns IrcMonitorStatus In-memory IRC monitor connection and per-channel join state (not persisted)
     * @throws ApiError
     */
    public static getIrcMonitorStatus(): CancelablePromise<IrcMonitorStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/irc-monitor/status',
        });
    }
    /**
     * Recorded stream sessions for monitored channels (newest first)
     * @returns RecordedStream Streams
     * @throws ApiError
     */
    public static listRecordedStreams({
        channelLogin,
        limit = 50,
        cursorStartedAt,
        cursorId,
    }: {
        /**
         * Filter by channel login
         */
        channelLogin?: string,
        limit?: number,
        /**
         * Keyset cursor; use with cursor_id from the last row of the previous page
         */
        cursorStartedAt?: string,
        cursorId?: number,
    }): CancelablePromise<Array<RecordedStream>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/streams',
            query: {
                'channel_login': channelLogin,
                'limit': limit,
                'cursor_started_at': cursorStartedAt,
                'cursor_id': cursorId,
            },
        });
    }
    /**
     * @returns RecordedStream Stream metadata
     * @throws ApiError
     */
    public static getRecordedStream({
        streamId,
    }: {
        streamId: number,
    }): CancelablePromise<RecordedStream> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/streams/{streamId}',
            path: {
                'streamId': streamId,
            },
            errors: {
                404: `Stream not found or channel not monitored`,
            },
        });
    }
    /**
     * Chat messages tagged with this stream (newest first)
     * @returns ChatHistoryEntry Messages
     * @throws ApiError
     */
    public static listRecordedStreamMessages({
        streamId,
        limit = 50,
        cursorCreatedAt,
        cursorId,
        username,
        text,
        chatterUserId,
    }: {
        streamId: number,
        limit?: number,
        cursorCreatedAt?: string,
        cursorId?: number,
        username?: string,
        text?: string,
        chatterUserId?: number,
    }): CancelablePromise<Array<ChatHistoryEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/streams/{streamId}/messages',
            path: {
                'streamId': streamId,
            },
            query: {
                'limit': limit,
                'cursor_created_at': cursorCreatedAt,
                'cursor_id': cursorId,
                'username': username,
                'text': text,
                'chatter_user_id': chatterUserId,
            },
            errors: {
                404: `Stream not found or channel not monitored`,
            },
        });
    }
    /**
     * Non-message activity in the stream time window (newest first)
     * @returns UserActivityEvent Activity events
     * @throws ApiError
     */
    public static listRecordedStreamActivity({
        streamId,
        limit = 50,
        cursorCreatedAt,
        cursorId,
    }: {
        streamId: number,
        limit?: number,
        cursorCreatedAt?: string,
        cursorId?: number,
    }): CancelablePromise<Array<UserActivityEvent>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/streams/{streamId}/activity',
            path: {
                'streamId': streamId,
            },
            query: {
                'limit': limit,
                'cursor_created_at': cursorCreatedAt,
                'cursor_id': cursorId,
            },
            errors: {
                404: `Stream not found or channel not monitored`,
            },
        });
    }
    /**
     * @returns StreamLeaderboardEntry Leaderboard rows
     * @throws ApiError
     */
    public static getRecordedStreamLeaderboard({
        streamId,
        sort,
        q,
    }: {
        streamId: number,
        sort?: StreamLeaderboardSort,
        /**
         * Filter chatter login (substring, case-insensitive)
         */
        q?: string,
    }): CancelablePromise<Array<StreamLeaderboardEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/streams/{streamId}/leaderboard',
            path: {
                'streamId': streamId,
            },
            query: {
                'sort': sort,
                'q': q,
            },
            errors: {
                404: `Stream not found or channel not monitored`,
            },
        });
    }
    /**
     * @returns UserActivityEvent Activity events (newest first)
     * @throws ApiError
     */
    public static listTwitchUserActivity({
        requestBody,
    }: {
        requestBody: ListTwitchUserActivityRequest,
    }): CancelablePromise<Array<UserActivityEvent>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/users/activity',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `User not found`,
            },
        });
    }
    /**
     * @returns ActivityTimelineSegment Merged chat presence intervals in the window
     * @throws ApiError
     */
    public static getTwitchUserActivityTimeline({
        requestBody,
    }: {
        requestBody: GetTwitchUserActivityTimelineRequest,
    }): CancelablePromise<Array<ActivityTimelineSegment>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/twitch/users/activity/timeline',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `User not found`,
            },
        });
    }
    /**
     * @returns AiSettings AI provider settings (API token is never returned in full)
     * @throws ApiError
     */
    public static getAiSettings(): CancelablePromise<AiSettings> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/ai/settings',
        });
    }
    /**
     * @returns AiSettings Updated settings
     * @throws ApiError
     */
    public static patchAiSettings({
        requestBody,
    }: {
        requestBody: PatchAiSettingsRequest,
    }): CancelablePromise<AiSettings> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/ai/settings',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns AiConversation Conversations (newest activity first)
     * @throws ApiError
     */
    public static listAiConversations(): CancelablePromise<Array<AiConversation>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/ai/conversations',
        });
    }
    /**
     * @returns AiConversation Created
     * @throws ApiError
     */
    public static createAiConversation({
        requestBody,
    }: {
        requestBody?: CreateAiConversationRequest,
    }): CancelablePromise<AiConversation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/ai/conversations',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static deleteAiConversation({
        conversationId,
    }: {
        conversationId: number,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/ai/conversations/{conversationId}',
            path: {
                'conversationId': conversationId,
            },
            errors: {
                404: `Not found`,
            },
        });
    }
    /**
     * @returns AiMessage Messages in order
     * @throws ApiError
     */
    public static listAiMessages({
        conversationId,
    }: {
        conversationId: number,
    }): CancelablePromise<Array<AiMessage>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/ai/conversations/{conversationId}/messages',
            path: {
                'conversationId': conversationId,
            },
            errors: {
                404: `Conversation not found`,
            },
        });
    }
    /**
     * @returns AiRunAccepted Message stored; agent run started asynchronously
     * @throws ApiError
     */
    public static createAiMessage({
        conversationId,
        requestBody,
    }: {
        conversationId: number,
        requestBody: CreateAiMessageRequest,
    }): CancelablePromise<AiRunAccepted> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/ai/conversations/{conversationId}/messages',
            path: {
                'conversationId': conversationId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Conversation not found`,
            },
        });
    }
    /**
     * @returns AiRunAccepted Confirmation recorded; agent resumes or rejects the tool call
     * @throws ApiError
     */
    public static confirmAiTool({
        conversationId,
        requestBody,
    }: {
        conversationId: number,
        requestBody: ConfirmAiToolRequest,
    }): CancelablePromise<AiRunAccepted> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/ai/conversations/{conversationId}/confirm',
            path: {
                'conversationId': conversationId,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Conversation or pending tool not found`,
            },
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    public static stopAiAgent({
        conversationId,
    }: {
        conversationId: number,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/ai/conversations/{conversationId}/stop',
            path: {
                'conversationId': conversationId,
            },
            errors: {
                404: `Conversation not found`,
            },
        });
    }
}
