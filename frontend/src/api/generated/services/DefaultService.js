import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * @returns LoginResponse Login success
     * @throws ApiError
     */
    static login({ requestBody, }) {
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
    static me() {
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
    static listTwitchUsers() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-users',
        });
    }
    /**
     * @returns TwitchUser Created
     * @throws ApiError
     */
    static createTwitchUser({ requestBody, }) {
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
    static updateTwitchUser({ requestBody, }) {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/settings/twitch-users/update',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                404: `Twitch user not found`,
            },
        });
    }
    /**
     * @returns string Blacklisted channel logins (lowercase)
     * @throws ApiError
     */
    static listChannelBlacklist() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/channel-blacklist',
        });
    }
    /**
     * @returns void
     * @throws ApiError
     */
    static setChannelBlacklist({ requestBody, }) {
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
    static getSuspicionSettings() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/suspicion-settings',
        });
    }
    /**
     * @returns SuspicionSettings Updated settings
     * @throws ApiError
     */
    static updateSuspicionSettings({ requestBody, }) {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/settings/suspicion-settings',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * @returns Rule Rule list
     * @throws ApiError
     */
    static listRules() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/rules',
        });
    }
    /**
     * @returns Rule Created
     * @throws ApiError
     */
    static createRule({ requestBody, }) {
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
    static countRules() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/rules/count',
        });
    }
    /**
     * @returns Rule Updated
     * @throws ApiError
     */
    static updateRule({ requestBody, }) {
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
    static deleteRule({ requestBody, }) {
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
     * @returns NotificationEntry Notification entries
     * @throws ApiError
     */
    static listNotifications() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/notifications',
        });
    }
    /**
     * @returns NotificationEntry Created
     * @throws ApiError
     */
    static createNotification({ requestBody, }) {
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
    static updateNotification({ requestBody, }) {
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
    static deleteNotification({ requestBody, }) {
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
    static listTwitchAccounts() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-accounts',
        });
    }
    /**
     * @returns TwitchAccount Created
     * @throws ApiError
     */
    static createTwitchAccount({ requestBody, }) {
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
    static countTwitchAccounts() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/settings/twitch-accounts/count',
        });
    }
    /**
     * @returns TwitchAccount Updated
     * @throws ApiError
     */
    static updateTwitchAccount({ requestBody, }) {
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
    static deleteTwitchAccount({ requestBody, }) {
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
    static startTwitchOAuth({ requestBody, }) {
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
    static sendMessage({ requestBody, }) {
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
    static listChatHistory({ channel, limit = 50, }) {
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
    static listTwitchMessages({ limit = 50, cursorCreatedAt, cursorId, username, text, channel, createdFrom, createdTo, chatterUserId, }) {
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
    static countTwitchMessages({ username, text, channel, createdFrom, createdTo, chatterUserId, }) {
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
     * @returns TwitchUser Users (marked first, then id desc)
     * @throws ApiError
     */
    static listTwitchDirectoryUsers({ username, limit = 50, cursorId, cursorMarked, }) {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/users',
            query: {
                'username': username,
                'limit': limit,
                'cursor_id': cursorId,
                'cursor_marked': cursorMarked,
            },
        });
    }
    /**
     * @returns CountResponse Count
     * @throws ApiError
     */
    static countTwitchDirectoryUsers({ username, }) {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/users/count',
            query: {
                'username': username,
            },
        });
    }
    /**
     * @returns TwitchUserProfile User profile
     * @throws ApiError
     */
    static getTwitchUserProfile({ requestBody, }) {
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
    static getChannelLive({ requestBody, }) {
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
    static listChannelChatters({ requestBody, }) {
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
    static getWatchUiHints() {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/twitch/watch/hints',
        });
    }
    /**
     * @returns IrcMonitorStatus In-memory IRC monitor connection and per-channel join state (not persisted)
     * @throws ApiError
     */
    static getIrcMonitorStatus() {
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
    static listRecordedStreams({ channelLogin, limit = 50, cursorStartedAt, cursorId, }) {
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
    static getRecordedStream({ streamId, }) {
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
    static listRecordedStreamMessages({ streamId, limit = 50, cursorCreatedAt, cursorId, username, text, chatterUserId, }) {
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
    static listRecordedStreamActivity({ streamId, limit = 50, cursorCreatedAt, cursorId, }) {
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
    static getRecordedStreamLeaderboard({ streamId, sort, q, }) {
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
    static listTwitchUserActivity({ requestBody, }) {
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
    static getTwitchUserActivityTimeline({ requestBody, }) {
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
}
