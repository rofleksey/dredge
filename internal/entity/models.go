package entity

import "time"

// ChatMessageListFilter selects persisted chat rows (newest first).
type ChatMessageListFilter struct {
	Username        string
	Text            string
	Channel         string
	StreamID        *int64
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	ChatterUserID   *int64
	Limit           int
	CursorCreatedAt *time.Time
	CursorID        *int64
}

// TwitchUserBrowseFilter drives directory search (keyset on marked DESC, id DESC).
type TwitchUserBrowseFilter struct {
	Username     string
	Limit        int
	CursorID     *int64
	CursorMarked *bool
}

// Account is an application login (admin or user role).
type Account struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
}

// Suspicion type values stored in twitch_users.sus_type.
const (
	SusTypeAutoAge       = "auto_age"
	SusTypeAutoBlacklist = "auto_blacklist"
	SusTypeAutoLowFollow = "auto_low_follows"
	SusTypeManual        = "manual"
)

// TwitchUserPatch is a partial update for twitch_users (nil fields are left unchanged).
type TwitchUserPatch struct {
	Monitored               *bool
	Marked                  *bool
	IsSus                   *bool
	SusType                 *string
	SusDescription          *string
	SusAutoSuppressed       *bool
	IrcOnlyWhenLive         *bool
	NotifyOffStreamMessages *bool
	NotifyStreamStart       *bool
}

// TwitchUser is a Twitch identity (e.g. a channel); monitored selects IRC join and keyword handling.
type TwitchUser struct {
	ID                      int64
	Username                string
	Monitored               bool
	Marked                  bool
	IsSus                   bool
	SusType                 *string
	SusDescription          *string
	SusAutoSuppressed       bool
	IrcOnlyWhenLive         bool
	NotifyOffStreamMessages bool
	NotifyStreamStart       bool
}

// Rule matches chat lines after user/channel allow/deny filters and regex.
type Rule struct {
	ID               int64
	Regex            string
	IncludedUsers    string
	DeniedUsers      string
	IncludedChannels string
	DeniedChannels   string
}

// TwitchAccount is an OAuth-linked Twitch identity used to send chat via Helix.
// ID is the Twitch user id (Helix "id" for the authorized user).
type TwitchAccount struct {
	ID           int64
	Username     string
	RefreshToken string
	AccountType  string
	CreatedAt    time.Time
}

// ChannelChatterEntry is one chatter in the IRC-maintained channel snapshot with presence and Helix meta.
type ChannelChatterEntry struct {
	Login            string
	UserTwitchID     int64
	PresentSince     time.Time
	AccountCreatedAt *time.Time
	MessageCount     *int64
}

// FollowedChannelRow is one outgoing follow edge stored from GQL sync.
type FollowedChannelRow struct {
	FollowedChannelID    int64
	FollowedChannelLogin string
	FollowedAt           *time.Time
}

// SuspicionSettings is the singleton row (id=1) driving automatic suspicion rules.
type SuspicionSettings struct {
	AutoCheckAccountAge bool
	AccountAgeSusDays   int
	AutoCheckBlacklist  bool
	AutoCheckLowFollows bool
	LowFollowsThreshold int
	MaxGQLFollowPages   int
}

// ChatHistoryMessage is a persisted Twitch chat line (IRC or sent via dredge).
type ChatHistoryMessage struct {
	ID                  int64
	Channel             string
	Username            string
	ChatterTwitchUserID *int64
	ChatterMarked       bool
	ChatterIsSus        bool
	Message             string
	KeywordMatch        bool
	MsgType             string
	BadgeTags           []string
	CreatedAt           time.Time
}

// UserActivityEventType is stored in user_activity_events.event_type.
const (
	UserActivityChatOnline  = "chat_online"
	UserActivityChatOffline = "chat_offline"
	UserActivityMessage     = "message"
)

// UserActivityEvent is a row for the activity feed / timeline.
type UserActivityEvent struct {
	ID                  int64
	ChatterTwitchUserID int64
	ChatterLogin        string
	EventType           string
	ChannelTwitchUserID *int64
	ChannelLogin        string
	Details             map[string]any
	CreatedAt           time.Time
}

// ActivityTimelineSegment is a merged chat presence interval for charts.
type ActivityTimelineSegment struct {
	ChannelTwitchUserID int64
	ChannelLogin        string
	Start               time.Time
	End                 time.Time
}

// UserActivityListFilter paginates activity for one chatter (newest first).
type UserActivityListFilter struct {
	ChatterUserID   int64
	Limit           int
	CursorCreatedAt *time.Time
	CursorID        *int64
}

// FollowedMonitoredChannel is follow metadata for a monitored broadcaster.
type FollowedMonitoredChannel struct {
	ChannelTwitchUserID int64
	ChannelLogin        string
	FollowedAt          *time.Time
}

// ChatterChannelPair is a distinct chatter + channel from chat history.
type ChatterChannelPair struct {
	ChatterID int64
	ChannelID int64
}

// NotificationEntry is a configured outbound notification (telegram, webhook, …).
type NotificationEntry struct {
	ID        int64
	Provider  string
	Settings  map[string]any
	Enabled   bool
	CreatedAt time.Time
}

// Stream is one recorded broadcast session for a monitored channel (Helix stream id).
type Stream struct {
	ID                  int64
	ChannelTwitchUserID int64
	ChannelLogin        string
	HelixStreamID       string
	StartedAt           time.Time
	EndedAt             *time.Time
	Title               string
	GameName            string
	CreatedAt           time.Time
}

// StreamListFilter lists streams for the Streams UI (newest first).
type StreamListFilter struct {
	ChannelLogin    string
	Limit           int
	CursorStartedAt *time.Time
	CursorID        *int64
}

// StreamLeaderboardSort is a query param for GET /twitch/streams/{id}/leaderboard.
type StreamLeaderboardSort string

const (
	StreamLeaderboardSortPresenceDesc StreamLeaderboardSort = "presence_desc"
	StreamLeaderboardSortPresenceAsc  StreamLeaderboardSort = "presence_asc"
	StreamLeaderboardSortMessagesDesc StreamLeaderboardSort = "messages_desc"
	StreamLeaderboardSortMessagesAsc  StreamLeaderboardSort = "messages_asc"
	StreamLeaderboardSortLoginAZ      StreamLeaderboardSort = "login_az"
	StreamLeaderboardSortLoginZA      StreamLeaderboardSort = "login_za"
	StreamLeaderboardSortAccountNew   StreamLeaderboardSort = "account_new"
	StreamLeaderboardSortAccountOld   StreamLeaderboardSort = "account_old"
)

// StreamLeaderboardRow is one aggregated chatter row for a stream.
type StreamLeaderboardRow struct {
	Login            string
	UserTwitchID     int64
	PresenceSeconds  int64
	MessageCount     int64
	AccountCreatedAt *time.Time
}

// UserActivityListFilterForStream paginates activity for a stream window (newest first).
type UserActivityListFilterForStream struct {
	ChannelTwitchUserID int64
	From                time.Time
	To                  time.Time
	Limit               int
	CursorCreatedAt     *time.Time
	CursorID            *int64
}
