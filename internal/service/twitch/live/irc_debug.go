package live

import (
	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"go.uber.org/zap"
)

func (r *Runtime) attachIRCMonitorDebug(client *twitchirc.Client) {
	client.OnConnect(func() {
		r.obs.Logger.Debug("irc monitor: on_connect")
	})
	client.OnWhisperMessage(func(m twitchirc.WhisperMessage) {
		r.obs.Logger.Debug("irc monitor: whisper", zap.String("user", m.User.Name))
	})
	client.OnClearChatMessage(func(m twitchirc.ClearChatMessage) {
		r.obs.Logger.Debug("irc monitor: clear_chat", zap.String("channel", m.Channel))
	})
	client.OnClearMessage(func(m twitchirc.ClearMessage) {
		r.obs.Logger.Debug("irc monitor: clear_msg", zap.String("channel", m.Channel))
	})
	client.OnRoomStateMessage(func(m twitchirc.RoomStateMessage) {
		r.obs.Logger.Debug("irc monitor: room_state", zap.String("channel", m.Channel))
	})
	client.OnUserNoticeMessage(func(m twitchirc.UserNoticeMessage) {
		r.obs.Logger.Debug("irc monitor: user_notice", zap.String("channel", m.Channel))
	})
	client.OnUserStateMessage(func(m twitchirc.UserStateMessage) {
		r.obs.Logger.Debug("irc monitor: user_state", zap.String("channel", m.Channel))
	})
	client.OnGlobalUserStateMessage(func(m twitchirc.GlobalUserStateMessage) {
		r.obs.Logger.Debug("irc monitor: global_user_state")
	})
	client.OnNoticeMessage(func(m twitchirc.NoticeMessage) {
		r.obs.Logger.Debug("irc monitor: notice", zap.String("channel", m.Channel))
	})
	client.OnPingMessage(func(m twitchirc.PingMessage) {
		r.obs.Logger.Debug("irc monitor: ping")
	})
	client.OnPongMessage(func(m twitchirc.PongMessage) {
		r.obs.Logger.Debug("irc monitor: pong")
	})
	client.OnPingSent(func() {
		r.obs.Logger.Debug("irc monitor: ping_sent")
	})
	client.OnReconnectMessage(func(m twitchirc.ReconnectMessage) {
		r.obs.Logger.Debug("irc monitor: reconnect")
	})
	client.OnNamesMessage(func(m twitchirc.NamesMessage) {
		r.obs.Logger.Debug("irc monitor: names", zap.String("channel", m.Channel))
	})
	client.OnUnsetMessage(func(m twitchirc.RawMessage) {
		r.obs.Logger.Debug("irc monitor: raw_unset", zap.String("raw", truncateString(m.Raw, 200)), zap.String("msg", truncateString(m.Message, 120)))
	})
}
