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
		toolFn(ToolCreateRule, "Create a new rule (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"name":            {Type: str},
				"enabled":         {Type: boolSchema},
				"event_type":      {Type: str, Description: "chat_message, stream_start, stream_end, interval"},
				"event_settings":  {Type: obj, Description: "JSON object, e.g. interval_seconds + channel for interval"},
				"middlewares":     {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: obj}},
				"action_type":     {Type: str, Description: "notify or send_chat"},
				"action_settings": {Type: obj},
				"use_shared_pool": {Type: boolSchema},
			},
			Required: []string{"name", "event_type", "event_settings", "middlewares", "action_type", "action_settings"},
		}),
		toolFn(ToolUpdateRule, "Update an existing rule by id (requires user approval).", jsonschema.Definition{
			Type: obj,
			Properties: map[string]jsonschema.Definition{
				"id":              {Type: integer},
				"name":            {Type: str},
				"enabled":         {Type: boolSchema},
				"event_type":      {Type: str},
				"event_settings":  {Type: obj},
				"middlewares":     {Type: jsonschema.Array, Items: &jsonschema.Definition{Type: obj}},
				"action_type":     {Type: str},
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
