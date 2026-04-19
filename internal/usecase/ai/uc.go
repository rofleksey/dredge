package ai

import (
	"context"
	"strings"
	"sync"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
	"github.com/rofleksey/dredge/internal/ws"
)

// Usecase persists AI settings and runs the chat agent.
type Usecase struct {
	repo    repository.Store
	tw      *twitchuc.Usecase
	rules   *rules.Usecase
	sett    *settings.Usecase
	hub *ws.Hub
	obs *observability.Stack

	runs sync.Map // int64 conversationID -> *runState
}

type runState struct {
	mu sync.Mutex

	cancel      context.CancelFunc
	pendingID   string
	approveCh   chan bool
}

// New constructs the AI usecase.
func New(repo repository.Store, tw *twitchuc.Usecase, rulesSvc *rules.Usecase, sett *settings.Usecase, hub *ws.Hub, obs *observability.Stack) *Usecase {
	return &Usecase{
		repo:  repo,
		tw:    tw,
		rules: rulesSvc,
		sett:  sett,
		hub:   hub,
		obs:   obs,
	}
}

func (u *Usecase) publicSettings(s entity.AISettings) (entity.AISettingsPublic, error) {
	tok := s.APIToken
	var last4 string
	if len(tok) >= 4 {
		last4 = tok[len(tok)-4:]
	}
	return entity.AISettingsPublic{
		BaseURL:    s.BaseURL,
		Model:      s.Model,
		HasToken:   strings.TrimSpace(tok) != "",
		TokenLast4: last4,
		UpdatedAt:  s.UpdatedAt,
	}, nil
}

// StopRun requests cancellation for an active agent run.
func (u *Usecase) StopRun(conversationID int64) {
	v, ok := u.runs.Load(conversationID)
	if !ok {
		return
	}
	st := v.(*runState)
	st.mu.Lock()
	if st.cancel != nil {
		st.cancel()
	}
	st.mu.Unlock()
}
