package settings

import (
	"context"
	"regexp"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"go.uber.org/zap"
)

func New(repo repository.Store, obs *observability.Stack) *Service {
	return &Service{repo: repo, obs: obs}
}

func (s *Service) ListTwitchUsers(ctx context.Context) ([]entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.list_twitch_users")
	defer span.End()

	out, err := s.repo.ListTwitchUsers(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list twitch users failed", err)
	}
	return out, err
}

func normalizeTwitchAccountLinkType(accountType string) string {
	if accountType == "bot" {
		return "bot"
	}
	return "main"
}

func (s *Service) CreateTwitchUser(ctx context.Context, id int64, username string) (entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.create_twitch_user")
	defer span.End()

	s.obs.Logger.Debug("create twitch user", zap.Int64("id", id), zap.String("username", username))

	out, err := s.repo.CreateTwitchUser(ctx, id, username)
	if err != nil {
		s.obs.LogError(ctx, span, "create twitch user failed", err,
			zap.Int64("id", id), zap.String("username", username))
	}

	return out, err
}

func (s *Service) PatchTwitchUser(ctx context.Context, id int64, patch entity.TwitchUserPatch) (entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.patch_twitch_user")
	defer span.End()

	cur, err := s.repo.GetTwitchUserByID(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch user load failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, err
	}

	// Turning off live-only implies off-stream IRC is off; apply before validating merged state.
	if patch.IrcOnlyWhenLive != nil && !*patch.IrcOnlyWhenLive {
		f := false
		patch.NotifyOffStreamMessages = &f
	}

	effIrcOnly := cur.IrcOnlyWhenLive
	if patch.IrcOnlyWhenLive != nil {
		effIrcOnly = *patch.IrcOnlyWhenLive
	}

	effNotifyOff := cur.NotifyOffStreamMessages
	if patch.NotifyOffStreamMessages != nil {
		effNotifyOff = *patch.NotifyOffStreamMessages
	}

	if effNotifyOff && !effIrcOnly {
		return entity.TwitchUser{}, entity.ErrInvalidTwitchUserMonitorSettings
	}

	out, err := s.repo.PatchTwitchUser(ctx, id, patch)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch user failed", err, zap.Int64("id", id))
	}

	return out, err
}

func (s *Service) ListChannelBlacklist(ctx context.Context) ([]string, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.list_channel_blacklist")
	defer span.End()

	out, err := s.repo.ListChannelBlacklist(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list channel blacklist failed", err)
	}
	return out, err
}

func (s *Service) SetChannelBlacklist(ctx context.Context, login string, add bool) error {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.set_channel_blacklist")
	defer span.End()

	var err error
	if add {
		err = s.repo.AddChannelBlacklist(ctx, login)
	} else {
		err = s.repo.RemoveChannelBlacklist(ctx, login)
	}

	if err != nil {
		s.obs.LogError(ctx, span, "set channel blacklist failed", err, zap.String("login", login), zap.Bool("add", add))
	}

	return err
}

func (s *Service) GetSuspicionSettings(ctx context.Context) (entity.SuspicionSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.get_suspicion_settings")
	defer span.End()

	out, err := s.repo.GetSuspicionSettings(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "get suspicion settings failed", err)
	}
	return out, err
}

func (s *Service) UpdateSuspicionSettings(ctx context.Context, in entity.SuspicionSettings) (entity.SuspicionSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.update_suspicion_settings")
	defer span.End()

	if err := s.repo.UpdateSuspicionSettings(ctx, in); err != nil {
		s.obs.LogError(ctx, span, "update suspicion settings failed", err)
		return entity.SuspicionSettings{}, err
	}

	return s.repo.GetSuspicionSettings(ctx)
}

func (s *Service) ListRules(ctx context.Context) ([]entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.list_rules")
	defer span.End()

	out, err := s.repo.ListRules(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list rules failed", err)
	}
	return out, err
}

func (s *Service) CountRules(ctx context.Context) (int64, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.count_rules")
	defer span.End()
	return s.repo.CountRules(ctx)
}

func (s *Service) CreateRule(ctx context.Context, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.create_rule")
	defer span.End()

	if _, err := regexp.Compile(r.Regex); err != nil {
		s.obs.LogError(ctx, span, "compile regex failed", err, zap.String("regex", r.Regex))
		return entity.Rule{}, err
	}

	out, err := s.repo.CreateRule(ctx, r)
	if err != nil {
		s.obs.LogError(ctx, span, "create rule failed", err, zap.String("regex", r.Regex))
	}

	return out, err
}

func (s *Service) UpdateRule(ctx context.Context, id int64, r entity.Rule) (entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.update_rule")
	defer span.End()

	if _, err := regexp.Compile(r.Regex); err != nil {
		s.obs.LogError(ctx, span, "compile regex failed", err, zap.String("regex", r.Regex))
		return entity.Rule{}, err
	}

	out, err := s.repo.UpdateRule(ctx, id, r)
	if err != nil {
		s.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", id))
	}

	return out, err
}

func (s *Service) DeleteRule(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.delete_rule")
	defer span.End()

	err := s.repo.DeleteRule(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", id))
	}
	return err
}

func (s *Service) ListNotifications(ctx context.Context) ([]entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.list_notifications")
	defer span.End()
	return s.repo.ListNotificationEntries(ctx)
}

func (s *Service) CreateNotification(ctx context.Context, provider string, settings map[string]any, enabled bool) (entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.create_notification")
	defer span.End()
	return s.repo.CreateNotificationEntry(ctx, provider, settings, enabled)
}

func (s *Service) UpdateNotification(ctx context.Context, id int64, provider *string, settings map[string]any, enabled *bool) (entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.update_notification")
	defer span.End()
	return s.repo.UpdateNotificationEntry(ctx, id, provider, settings, enabled)
}

func (s *Service) DeleteNotification(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.delete_notification")
	defer span.End()
	return s.repo.DeleteNotificationEntry(ctx, id)
}

func (s *Service) ListTwitchAccounts(ctx context.Context) ([]entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.list_twitch_accounts")
	defer span.End()

	out, err := s.repo.ListTwitchAccounts(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list twitch accounts failed", err)
	}

	return out, err
}

func (s *Service) CountTwitchAccounts(ctx context.Context) (int64, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.count_twitch_accounts")
	defer span.End()
	return s.repo.CountTwitchAccounts(ctx)
}

func (s *Service) CreateTwitchAccount(ctx context.Context, id int64, username, refreshToken, accountType string) (entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.create_twitch_account")
	defer span.End()

	accountType = normalizeTwitchAccountLinkType(accountType)

	s.obs.Logger.Debug("create twitch account", zap.Int64("id", id), zap.String("username", username), zap.String("account_type", accountType))

	out, err := s.repo.CreateTwitchAccount(ctx, id, username, refreshToken, accountType)
	if err != nil {
		s.obs.LogError(ctx, span, "create twitch account failed", err, zap.String("username", username))
	}

	return out, err
}

func (s *Service) PatchTwitchAccount(ctx context.Context, id int64, accountType *string) (entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.patch_twitch_account")
	defer span.End()

	if accountType != nil {
		t := normalizeTwitchAccountLinkType(*accountType)
		accountType = &t
	}

	out, err := s.repo.PatchTwitchAccount(ctx, id, accountType)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch account failed", err, zap.Int64("id", id))
	}

	return out, err
}

func (s *Service) DeleteTwitchAccount(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "service.settings.delete_twitch_account")
	defer span.End()

	err := s.repo.DeleteTwitchAccount(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "delete twitch account failed", err, zap.Int64("id", id))
	}
	return err
}
