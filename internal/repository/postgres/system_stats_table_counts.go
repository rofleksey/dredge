package postgres

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (r *Repository) SystemStatsTableCounts(ctx context.Context) (entity.SystemStatsTableCounts, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.system_stats_table_counts")
	defer span.End()

	const q = `
SELECT
	(SELECT COUNT(*)::bigint FROM twitch_users),
	(SELECT COUNT(*)::bigint FROM twitch_accounts WHERE deleted_at IS NULL),
	(SELECT COUNT(*)::bigint FROM twitch_accounts),
	(SELECT COUNT(*)::bigint FROM rules),
	(SELECT COUNT(*)::bigint FROM notification_entries),
	(SELECT COUNT(*)::bigint FROM streams),
	(SELECT COUNT(*)::bigint FROM streams WHERE ended_at IS NULL),
	(SELECT COUNT(*)::bigint FROM chat_messages),
	(SELECT COUNT(*)::bigint FROM channel_chatters),
	(SELECT COUNT(*)::bigint FROM user_activity_events),
	(SELECT COUNT(*)::bigint FROM twitch_user_helix_meta),
	(SELECT COUNT(*)::bigint FROM twitch_user_channel_follows),
	(SELECT COUNT(*)::bigint FROM user_followed_channels),
	(SELECT COUNT(*)::bigint FROM channel_blacklist),
	(SELECT COUNT(*)::bigint FROM rule_trigger_events),
	(SELECT COUNT(*)::bigint FROM irc_joined_samples),
	(SELECT COUNT(*)::bigint FROM twitch_discovery_candidates),
	(SELECT COUNT(*)::bigint FROM twitch_discovery_denied),
	(SELECT COUNT(*)::bigint FROM ai_conversations),
	(SELECT COUNT(*)::bigint FROM ai_messages)
`

	row := r.pool.QueryRow(ctx, q)

	var out entity.SystemStatsTableCounts

	err := row.Scan(
		&out.TwitchUsers,
		&out.TwitchAccountsActive,
		&out.TwitchAccountsAll,
		&out.Rules,
		&out.NotificationEntries,
		&out.Streams,
		&out.StreamsOpen,
		&out.ChatMessages,
		&out.ChannelChatters,
		&out.UserActivityEvents,
		&out.TwitchUserHelixMeta,
		&out.TwitchUserChannelFollows,
		&out.UserFollowedChannels,
		&out.ChannelBlacklist,
		&out.RuleTriggerEvents,
		&out.IrcJoinedSamples,
		&out.TwitchDiscoveryCandidates,
		&out.TwitchDiscoveryDenied,
		&out.AiConversations,
		&out.AiMessages,
	)
	if err != nil {
		r.obs.LogError(ctx, span, "system stats table counts query failed", err)
		return entity.SystemStatsTableCounts{}, err
	}

	return out, nil
}
