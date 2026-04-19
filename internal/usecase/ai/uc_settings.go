package ai

import (
	"context"
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

// GetAISettings loads settings from DB.
func (u *Usecase) GetAISettings(ctx context.Context) (entity.AISettings, entity.AISettingsPublic, error) {
	s, err := u.repo.GetAISettings(ctx)
	if err != nil {
		return entity.AISettings{}, entity.AISettingsPublic{}, err
	}
	pub, err := u.publicSettings(s)
	if err != nil {
		return entity.AISettings{}, entity.AISettingsPublic{}, err
	}
	return s, pub, nil
}

// PatchAISettings merges fields; when newToken is non-empty it replaces the stored API token.
func (u *Usecase) PatchAISettings(ctx context.Context, baseURL, model *string, newToken *string) (entity.AISettingsPublic, error) {
	s, err := u.repo.GetAISettings(ctx)
	if err != nil {
		return entity.AISettingsPublic{}, err
	}
	if baseURL != nil {
		s.BaseURL = strings.TrimSpace(*baseURL)
	}
	if model != nil {
		s.Model = strings.TrimSpace(*model)
	}
	if newToken != nil && strings.TrimSpace(*newToken) != "" {
		s.APIToken = *newToken
	}
	if err := u.repo.UpsertAISettings(ctx, s); err != nil {
		return entity.AISettingsPublic{}, err
	}
	s2, err := u.repo.GetAISettings(ctx)
	if err != nil {
		return entity.AISettingsPublic{}, err
	}
	return u.publicSettings(s2)
}
