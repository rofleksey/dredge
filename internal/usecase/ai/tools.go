package ai

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Tool names used with the LLM (snake_case).
const (
	ToolListTwitchMessages           = "list_twitch_messages"
	ToolGetTwitchUserProfile         = "get_twitch_user_profile"
	ToolListTwitchUserActivity       = "list_twitch_user_activity"
	ToolGetTwitchUserActivityTimeline = "get_twitch_user_activity_timeline"
	ToolListTwitchDirectoryUsers     = "list_twitch_directory_users"
	ToolListChatHistory              = "list_chat_history"
	ToolListRules                    = "list_rules"
	ToolRuleTemplateVariables        = "rule_template_variables"
	ToolTestRuleRegex                = "test_rule_regex"
	ToolListTwitchUsers              = "list_twitch_users"
	ToolListNotifications            = "list_notifications"
	ToolListChannelBlacklist         = "list_channel_blacklist"
	ToolGetSuspicionSettings         = "get_suspicion_settings"
	ToolGetIrcMonitorSettings        = "get_irc_monitor_settings"
	ToolListTwitchAccounts           = "list_twitch_accounts"
	ToolCreateRule                   = "create_rule"
	ToolUpdateRule                   = "update_rule"
	ToolDeleteRule                   = "delete_rule"
	ToolSendTwitchMessage            = "send_twitch_message"
	ToolCountTwitchMessages          = "count_twitch_messages"
	ToolCountTwitchDirectoryUsers    = "count_twitch_directory_users"
	ToolGetChannelLive               = "get_channel_live"
	ToolListChannelChatters          = "list_channel_chatters"
	ToolGetIrcMonitorStatus          = "get_irc_monitor_status"
	ToolGetWatchUiHints              = "get_watch_ui_hints"
	ToolListMonitoredStreams         = "list_monitored_streams"
	ToolGetMonitoredStream           = "get_monitored_stream"
	ToolListStreamMessages           = "list_stream_messages"
	ToolListStreamActivity           = "list_stream_activity"
	ToolGetStreamLeaderboard         = "get_stream_leaderboard"
	ToolGetTwitchUser                = "get_twitch_user"
	ToolCountRules                   = "count_rules"
	ToolCreateNotification           = "create_notification"
	ToolUpdateNotification           = "update_notification"
	ToolDeleteNotification           = "delete_notification"
	ToolSetChannelBlacklist          = "set_channel_blacklist"
	ToolUpdateSuspicionSettings      = "update_suspicion_settings"
	ToolUpdateIrcMonitorSettings     = "update_irc_monitor_settings"
	ToolCreateTwitchUser             = "create_twitch_user"
	ToolPatchTwitchUser              = "patch_twitch_user"
	ToolCreateTwitchAccount          = "create_twitch_account"
	ToolPatchTwitchAccount           = "patch_twitch_account"
	ToolDeleteTwitchAccount          = "delete_twitch_account"
)

var readOnlyTools = map[string]struct{}{
	ToolListTwitchMessages:            {},
	ToolGetTwitchUserProfile:          {},
	ToolListTwitchUserActivity:        {},
	ToolGetTwitchUserActivityTimeline: {},
	ToolListTwitchDirectoryUsers:      {},
	ToolListChatHistory:               {},
	ToolListRules:                     {},
	ToolRuleTemplateVariables:         {},
	ToolTestRuleRegex:                 {},
	ToolListTwitchUsers:               {},
	ToolListNotifications:             {},
	ToolListChannelBlacklist:          {},
	ToolGetSuspicionSettings:          {},
	ToolGetIrcMonitorSettings:         {},
	ToolListTwitchAccounts:            {},
	ToolCountTwitchMessages:           {},
	ToolCountTwitchDirectoryUsers:     {},
	ToolGetChannelLive:                {},
	ToolListChannelChatters:           {},
	ToolGetIrcMonitorStatus:           {},
	ToolGetWatchUiHints:               {},
	ToolListMonitoredStreams:          {},
	ToolGetMonitoredStream:            {},
	ToolListStreamMessages:            {},
	ToolListStreamActivity:            {},
	ToolGetStreamLeaderboard:          {},
	ToolGetTwitchUser:                 {},
	ToolCountRules:                    {},
}

// ToolIsReadOnly returns true when the tool may run without user confirmation.
func ToolIsReadOnly(name string) bool {
	_, ok := readOnlyTools[name]
	return ok
}

// LLMTools returns OpenAI tool definitions for chat completions.
func LLMTools() []openai.Tool {
	str := jsonschema.String
	integer := jsonschema.Integer
	boolSchema := jsonschema.Boolean
	obj := jsonschema.Object

	return []openai.Tool{
		toolFn(ToolListTwitchMessages, "Search persisted chat messages (newest first).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"username":     {Type: str, Description: "Filter by chatter login (substring)"},
				"channel":      {Type: str, Description: "Filter by channel login"},
				"text":         {Type: str, Description: "Filter by message body substring"},
				"limit":        {Type: integer, Description: "1-200, default 50"},
				"chatter_user_id": {Type: integer, Description: "Twitch user id of chatter when known"},
			},
		}),
		toolFn(ToolGetTwitchUserProfile, "Profile, stats, follows, and blacklist context for a Twitch user id.", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolListTwitchUserActivity, "Activity events for a user (newest first).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":    {Type: integer},
				"limit": {Type: integer},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolGetTwitchUserActivityTimeline, "Merged chat presence intervals for a user in a time window.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":   {Type: integer},
				"from": {Type: str, Description: "RFC3339, optional"},
				"to":   {Type: str, Description: "RFC3339, optional"},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolListTwitchDirectoryUsers, "Browse known Twitch users (directory).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"username":       {Type: str},
				"limit":          {Type: integer},
				"monitored_only": {Type: boolSchema},
			},
		}),
		toolFn(ToolListChatHistory, "Recent chat lines for one channel (IRC history).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"channel": {Type: str},
				"limit":   {Type: integer},
			},
			Required: []string{"channel"},
		}),
		toolFn(ToolListRules, "List automation rules.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolRuleTemplateVariables, "Describe $PLACEHOLDER variables for rule message templates.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolTestRuleRegex, "Test a regex pattern against a sample (safe, no side effects).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"pattern":           {Type: str},
				"sample":            {Type: str},
				"case_insensitive":  {Type: boolSchema},
			},
			Required: []string{"pattern", "sample"},
		}),
		toolFn(ToolListTwitchUsers, "List configured Twitch channels/users in settings.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"monitored_only": {Type: boolSchema},
			},
		}),
		toolFn(ToolListNotifications, "List notification provider entries.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolListChannelBlacklist, "List globally blacklisted channel logins.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolGetSuspicionSettings, "Get automatic suspicion thresholds.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolGetIrcMonitorSettings, "Get IRC monitor OAuth identity settings.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolListTwitchAccounts, "List linked Twitch OAuth accounts.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolCountTwitchMessages, "Count persisted chat messages matching optional filters (same fields as list_twitch_messages plus created_from, created_to RFC3339).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"username":         {Type: str},
				"text":             {Type: str},
				"channel":          {Type: str},
				"chatter_user_id":  {Type: integer},
				"created_from":     {Type: str, Description: "RFC3339"},
				"created_to":       {Type: str, Description: "RFC3339"},
			},
		}),
		toolFn(ToolCountTwitchDirectoryUsers, "Count directory users (optional username substring, monitored_only).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"username":        {Type: str},
				"monitored_only":  {Type: boolSchema},
			},
		}),
		toolFn(ToolGetChannelLive, "Helix live status and chatter snapshot for a channel login.", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"login": {Type: str}},
			Required:   []string{"login"},
		}),
		toolFn(ToolListChannelChatters, "IRC snapshot chatters for a channel; requires linked twitch account_id and channel login.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"account_id":          {Type: integer, Description: "twitch_accounts.id"},
				"login":               {Type: str, Description: "Channel login"},
				"session_started_at":  {Type: str, Description: "RFC3339, optional, scopes message counts"},
			},
			Required: []string{"account_id", "login"},
		}),
		toolFn(ToolGetIrcMonitorStatus, "Whether the IRC monitor is connected and per-channel join status.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolGetWatchUiHints, "UI poll intervals in seconds (viewer, channel chatters, monitored live).", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolListMonitoredStreams, "List recorded/monitored streams (newest first). Optional channel_login, limit, cursor_started_at+cursor_id.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"channel_login":      {Type: str},
				"limit":              {Type: integer},
				"cursor_started_at":  {Type: str, Description: "RFC3339"},
				"cursor_id":          {Type: integer},
			},
		}),
		toolFn(ToolGetMonitoredStream, "Get one monitored stream by id.", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolListStreamMessages, "Chat messages tagged with a stream id.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"stream_id":          {Type: integer},
				"limit":              {Type: integer},
				"username":           {Type: str},
				"text":               {Type: str},
				"chatter_user_id":    {Type: integer},
				"cursor_created_at":  {Type: str, Description: "RFC3339"},
				"cursor_id":          {Type: integer},
			},
			Required: []string{"stream_id"},
		}),
		toolFn(ToolListStreamActivity, "Non-message activity for a stream time window.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"stream_id":          {Type: integer},
				"limit":              {Type: integer},
				"cursor_created_at":  {Type: str, Description: "RFC3339"},
				"cursor_id":          {Type: integer},
			},
			Required: []string{"stream_id"},
		}),
		toolFn(ToolGetStreamLeaderboard, "Aggregated leaderboard for a stream. sort: presence_desc, presence_asc, messages_desc, messages_asc, login_az, login_za, account_new, account_old.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"stream_id": {Type: integer},
				"sort":      {Type: str},
				"q":         {Type: str, Description: "Filter logins"},
			},
			Required: []string{"stream_id"},
		}),
		toolFn(ToolGetTwitchUser, "Load twitch_users row by Twitch user id.", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolCountRules, "Count automation rules.", jsonschema.Definition{Type: obj, Properties: map[string]jsonschema.Definition{}}),
		toolFn(ToolCreateRule, "Create a new rule (requires user approval). event_type: chat_message | stream_start | stream_end | interval. action_type: notify | send_chat. middleware type: filter_channel | filter_user | match_regex | contains_word | cooldown. Use list_rules and rule_template_variables before editing.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"name":            {Type: str},
				"enabled":         {Type: boolSchema},
				"event_type":      {Type: str, Description: "chat_message | stream_start | stream_end | interval"},
				"event_settings":  {Type: obj, Description: "For interval: interval_seconds (int), channel (login)."},
				"middlewares":     {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: obj, Description: "{type, settings}"}},
				"action_type":     {Type: str, Description: "notify | send_chat"},
				"action_settings": {Type: obj},
				"use_shared_pool": {Type: boolSchema},
			},
			Required: []string{"name", "event_type", "event_settings", "middlewares", "action_type", "action_settings"},
		}),
		toolFn(ToolUpdateRule, "Patch an existing rule by id (requires user approval). Only include fields to change; omitted fields keep DB values. event_settings/action_settings merge shallowly with existing maps.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":              {Type: integer},
				"name":            {Type: str},
				"enabled":         {Type: boolSchema},
				"event_type":      {Type: str, Description: "chat_message | stream_start | stream_end | interval"},
				"event_settings":  {Type: obj},
				"middlewares":     {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: obj}},
				"action_type":     {Type: str, Description: "notify | send_chat"},
				"action_settings": {Type: obj},
				"use_shared_pool": {Type: boolSchema},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolDeleteRule, "Delete a rule by id (requires user approval).", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolCreateNotification, "Create a notification entry (requires user approval). provider e.g. telegram, webhook.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"provider": {Type: str},
				"settings": {Type: obj},
				"enabled":  {Type: boolSchema},
			},
			Required: []string{"provider", "settings"},
		}),
		toolFn(ToolUpdateNotification, "Update notification by id (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":       {Type: integer},
				"provider": {Type: str},
				"settings": {Type: obj},
				"enabled":  {Type: boolSchema},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolDeleteNotification, "Delete notification by id (requires user approval).", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolSetChannelBlacklist, "Add or remove a channel from the global blacklist (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"login": {Type: str},
				"add":   {Type: boolSchema, Description: "true to add, false to remove"},
			},
			Required: []string{"login", "add"},
		}),
		toolFn(ToolUpdateSuspicionSettings, "Replace suspicion automation settings (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"auto_check_account_age": {Type: boolSchema},
				"account_age_sus_days":   {Type: integer},
				"auto_check_blacklist":   {Type: boolSchema},
				"auto_check_low_follows": {Type: boolSchema},
				"low_follows_threshold":  {Type: integer},
				"max_gql_follow_pages":   {Type: integer},
			},
			Required: []string{"auto_check_account_age", "account_age_sus_days", "auto_check_blacklist", "auto_check_low_follows", "low_follows_threshold", "max_gql_follow_pages"},
		}),
		toolFn(ToolUpdateIrcMonitorSettings, "Update IRC monitor OAuth account and enrichment cooldown (requires user approval). Omit a field to keep current; set oauth_twitch_account_id to null for anonymous IRC.", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"oauth_twitch_account_id":   {Type: integer, Nullable: true, Description: "Linked twitch_accounts.id or null"},
				"enrichment_cooldown_hours": {Type: integer},
			},
		}),
		toolFn(ToolCreateTwitchUser, "Resolve a Twitch channel by name and add it to monitored users (requires user approval).", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"name": {Type: str, Description: "Channel login"}},
			Required:   []string{"name"},
		}),
		toolFn(ToolPatchTwitchUser, "Partially update a twitch_users row by id (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":                         {Type: integer},
				"monitored":                  {Type: boolSchema},
				"marked":                     {Type: boolSchema},
				"is_sus":                     {Type: boolSchema},
				"sus_type":                   {Type: str},
				"sus_description":            {Type: str},
				"sus_auto_suppressed":        {Type: boolSchema},
				"irc_only_when_live":         {Type: boolSchema},
				"notify_off_stream_messages": {Type: boolSchema},
				"notify_stream_start":        {Type: boolSchema},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolCreateTwitchAccount, "Insert a linked Twitch OAuth account row (requires user approval; usually for automation).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":            {Type: integer, Description: "Twitch user id"},
				"username":      {Type: str},
				"refresh_token": {Type: str},
				"account_type":  {Type: str, Description: "main or bot"},
			},
			Required: []string{"id", "username", "refresh_token"},
		}),
		toolFn(ToolPatchTwitchAccount, "Update linked account type (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":           {Type: integer},
				"account_type": {Type: str},
			},
			Required: []string{"id"},
		}),
		toolFn(ToolDeleteTwitchAccount, "Delete a linked Twitch OAuth account by id (requires user approval).", jsonschema.Definition{
			Type:       obj,
			Properties: map[string]jsonschema.Definition{"id": {Type: integer}},
			Required:   []string{"id"},
		}),
		toolFn(ToolSendTwitchMessage, "Send a chat message via Helix using a linked Twitch account (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"account_id": {Type: integer, Description: "Linked twitch_accounts.id"},
				"channel":    {Type: str},
				"message":    {Type: str},
			},
			Required: []string{"account_id", "channel", "message"},
		}),
	}
}

func toolFn(name, description string, params jsonschema.Definition) openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters:  params,
		},
	}
}
