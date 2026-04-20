package repository

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

// Store is the persistence contract shared by application services.
type Store interface {
	ListTwitchUsers(ctx context.Context) ([]entity.TwitchUser, error)
	CreateTwitchUser(ctx context.Context, id int64, username string) (entity.TwitchUser, error)
	PatchTwitchUser(ctx context.Context, id int64, patch entity.TwitchUserPatch) (entity.TwitchUser, error)
	ListMonitoredTwitchUsers(ctx context.Context) ([]entity.TwitchUser, error)
	ListMonitoredOrMarkedTwitchUserIDs(ctx context.Context) ([]int64, error)

	ListRules(ctx context.Context) ([]entity.Rule, error)
	ListRuleTriggerEvents(ctx context.Context, f entity.RuleTriggerListFilter) ([]entity.RuleTriggerEvent, error)
	CountRules(ctx context.Context) (int64, error)
	CreateRule(ctx context.Context, r entity.Rule) (entity.Rule, error)
	UpdateRule(ctx context.Context, id int64, r entity.Rule) (entity.Rule, error)
	DeleteRule(ctx context.Context, id int64) error

	ListNotificationEntries(ctx context.Context, f entity.NotificationListFilter) ([]entity.NotificationEntry, error)
	ListEnabledNotificationEntries(ctx context.Context) ([]entity.NotificationEntry, error)
	CreateNotificationEntry(ctx context.Context, provider string, settings map[string]any, enabled bool) (entity.NotificationEntry, error)
	UpdateNotificationEntry(ctx context.Context, id int64, provider *string, settings map[string]any, enabled *bool) (entity.NotificationEntry, error)
	DeleteNotificationEntry(ctx context.Context, id int64) error

	ListTwitchAccounts(ctx context.Context) ([]entity.TwitchAccount, error)
	CountTwitchAccounts(ctx context.Context) (int64, error)
	CreateTwitchAccount(ctx context.Context, id int64, username, refreshToken, accountType string) (entity.TwitchAccount, error)
	PatchTwitchAccount(ctx context.Context, id int64, accountType *string) (entity.TwitchAccount, error)
	DeleteTwitchAccount(ctx context.Context, id int64) error
	GetTwitchAccountByID(ctx context.Context, id int64) (entity.TwitchAccount, error)
	GetTwitchAccountByTwitchUserID(ctx context.Context, twitchUserID int64) (entity.TwitchAccount, error)
	UpdateTwitchRefreshToken(ctx context.Context, id int64, refreshToken string) error

	InsertChatMessage(ctx context.Context, channelTwitchUserID int64, chatterTwitchUserID *int64, chatterUsername, body string, keywordMatch bool, msgType string, badgeTags []string, firstMessage bool) (int64, error)
	InsertChatMessageForChannelLogin(ctx context.Context, channelLogin string, chatterTwitchUserID *int64, chatterUsername, body string, keywordMatch bool, msgType string, badgeTags []string, firstMessage bool) (int64, error)
	UpsertTwitchUserFromChat(ctx context.Context, id int64, username string) (inserted bool, err error)
	IsMonitoredChannel(ctx context.Context, channel string) (bool, error)
	// MonitoredChannelTwitchUserID returns the twitch_users.id for a monitored channel by login (ok=false if not monitored).
	MonitoredChannelTwitchUserID(ctx context.Context, channel string) (id int64, ok bool, err error)
	ListChatHistory(ctx context.Context, channel string, limit int) ([]entity.ChatHistoryMessage, error)
	ListChatMessages(ctx context.Context, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error)
	CountChatMessages(ctx context.Context, f entity.ChatMessageListFilter) (int64, error)
	ListTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) ([]entity.TwitchDirectoryEntry, error)
	CountTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) (int64, error)
	GetTwitchUserByID(ctx context.Context, id int64) (entity.TwitchUser, error)
	CountChatMessagesByChatter(ctx context.Context, chatterID int64) (int64, error)
	IsTwitchUserMarked(ctx context.Context, id int64) (bool, error)
	IsTwitchUserSuspicious(ctx context.Context, id int64) (bool, error)

	ReplaceChannelChattersSnapshot(ctx context.Context, channelTwitchUserID int64, chatterIDs []int64) error
	UpsertChannelChatterPresence(ctx context.Context, channelTwitchUserID, chatterTwitchUserID int64) (presentSince time.Time, err error)
	DeleteChannelChatterPresence(ctx context.Context, channelTwitchUserID, chatterTwitchUserID int64) (presentSince time.Time, deleted bool, err error)
	ListChannelChatterIDs(ctx context.Context, channelTwitchUserID int64) ([]int64, error)
	CountChannelChatters(ctx context.Context, channelTwitchUserID int64) (int64, error)
	ListChannelChatterEntries(ctx context.Context, channelTwitchUserID int64) ([]entity.ChannelChatterEntry, error)
	InsertUserActivityEvent(ctx context.Context, chatterID int64, eventType string, channelTwitchUserID *int64, details map[string]any) error
	ListUserActivityEvents(ctx context.Context, f entity.UserActivityListFilter) ([]entity.UserActivityEvent, error)
	ListUserActivityEventsForTimeline(ctx context.Context, chatterID int64, from, to time.Time) ([]entity.UserActivityEvent, error)
	UpsertHelixMeta(ctx context.Context, twitchUserID int64, accountCreatedAt *time.Time, profileImageURL *string, fetchedAt time.Time) error
	GetHelixMeta(ctx context.Context, twitchUserID int64) (accountCreatedAt *time.Time, helixFetchedAt *time.Time, profileImageURL *string, err error)
	UpsertChannelFollow(ctx context.Context, chatterID, channelID int64, followedAt *time.Time, checkedAt time.Time) error
	ListFollowedMonitoredChannels(ctx context.Context, chatterID int64) ([]entity.FollowedMonitoredChannel, error)
	ListDistinctChattersWithMessages(ctx context.Context, limit int) ([]int64, error)
	ListChatterChannelPairsForFollowEnrichment(ctx context.Context, limit int) ([]entity.ChatterChannelPair, error)
	TruncateChannelChatters(ctx context.Context) error
	TwitchUserIDByUsername(ctx context.Context, username string) (int64, error)

	ActiveStreamIDForChannel(ctx context.Context, channelTwitchUserID int64) (*int64, error)
	UpsertStreamFromHelix(ctx context.Context, channelTwitchUserID int64, helixStreamID string, startedAt time.Time, title, gameName string, viewerCount *int64) (int64, error)
	CloseOpenStreamsForChannel(ctx context.Context, channelTwitchUserID int64) error
	GetStreamByID(ctx context.Context, id int64) (entity.Stream, error)
	GetMonitoredStreamByID(ctx context.Context, id int64) (entity.Stream, error)
	ListMonitoredStreams(ctx context.Context, f entity.StreamListFilter) ([]entity.Stream, error)
	CountChatMessagesPerChatterForStream(ctx context.Context, streamID int64) (map[int64]int64, error)
	CountChatMessagesPerChatterForChannelSince(ctx context.Context, channelTwitchUserID int64, since time.Time) (map[int64]int64, error)
	ListUserActivityEventsForChannelPresence(ctx context.Context, channelTwitchUserID int64, from, to time.Time) ([]entity.UserActivityEvent, error)
	ListUserActivityForStream(ctx context.Context, f entity.UserActivityListFilterForStream) ([]entity.UserActivityEvent, error)

	ReplaceUserFollowedChannels(ctx context.Context, followerID int64, rows []entity.FollowedChannelRow) error
	ListUserFollowedChannels(ctx context.Context, followerID int64) ([]entity.FollowedChannelRow, error)
	ListChannelBlacklist(ctx context.Context) ([]string, error)
	AddChannelBlacklist(ctx context.Context, login string) error
	RemoveChannelBlacklist(ctx context.Context, login string) error
	GetSuspicionSettings(ctx context.Context) (entity.SuspicionSettings, error)
	UpdateSuspicionSettings(ctx context.Context, s entity.SuspicionSettings) error
	GetIrcMonitorSettings(ctx context.Context) (entity.IrcMonitorSettings, error)
	UpdateIrcMonitorSettings(ctx context.Context, s entity.IrcMonitorSettings) error
	InsertIrcJoinedSample(ctx context.Context, joinedCount int) error
	InsertRuleTriggerEvent(ctx context.Context, ruleID int64, ruleName, triggerEvent, actionType, displayText string) error
	ListIrcJoinedSamples(ctx context.Context, from, to time.Time) ([]entity.IrcJoinedSample, error)
	ListLinkedTwitchAccountUserIDs(ctx context.Context) ([]int64, error)

	GetAISettings(ctx context.Context) (entity.AISettings, error)
	UpsertAISettings(ctx context.Context, s entity.AISettings) error
	ListAIConversations(ctx context.Context) ([]entity.AIConversation, error)
	CreateAIConversation(ctx context.Context, title *string) (entity.AIConversation, error)
	GetAIConversation(ctx context.Context, id int64) (entity.AIConversation, error)
	DeleteAIConversation(ctx context.Context, id int64) error
	TouchAIConversation(ctx context.Context, id int64) error
	ListAIMessages(ctx context.Context, conversationID int64) ([]entity.AIMessage, error)
	InsertAIMessage(ctx context.Context, m entity.AIMessage) (entity.AIMessage, error)
	SetAIMessageMetadata(ctx context.Context, messageID int64, metadata map[string]any) error
}
