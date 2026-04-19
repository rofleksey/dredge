package app

import (
	"context"

	"go.uber.org/fx"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func newRulesServices(
	repo repository.Store,
	obs *observability.Stack,
	tw *twitchuc.Usecase,
) (*rules.Engine, *rules.Usecase, error) {
	eng := rules.NewEngine(rules.Config{
		Repo:           repo,
		Helix:          tw.Client,
		Notify:         tw.LiveRuntime(),
		Send:           tw,
		PersistContext: func() context.Context { return tw.PersistContext() },
		Obs:            obs,
	})

	svc := rules.NewUsecase(repo, obs, eng, tw.RestartMonitor)

	return eng, svc, nil
}

func registerRulesLifecycle(lc fx.Lifecycle, eng *rules.Engine, svc *rules.Usecase, tw *twitchuc.Usecase) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			tw.LiveRuntime().SetRuleEngine(eng)

			eng.Start(context.Background())

			if err := svc.Bootstrap(ctx); err != nil {
				return err
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			tw.LiveRuntime().SetRuleEngine(nil)

			eng.Stop()

			return nil
		},
	})
}
