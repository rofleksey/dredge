package rules

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

const (
	defaultWorkerCount = 8
	workQueueSize      = 256
	schedulerTick      = time.Second
)

// Config wires the rule engine.
type Config struct {
	Repo           repository.Store
	Helix          *helix.Client
	Notify         NotifyDispatcher
	Send           SendMessenger
	PersistContext func() context.Context
	Obs            *observability.Stack
}

// Engine evaluates rules with a worker pool and interval scheduler.
type Engine struct {
	deps     evalDeps
	persist  func() context.Context
	obs      *observability.Stack
	notify   NotifyDispatcher
	send     SendMessenger
	cooldown *cooldownTracker

	rules atomic.Value // []entity.Rule

	work chan workItem

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	intervalMu   sync.Mutex
	intervalNext map[int64]time.Time
}

type workItem struct {
	rule    entity.Rule
	payload EvalPayload
}

// NewEngine constructs an engine; call Start after construction.
func NewEngine(cfg Config) *Engine {
	pc := cfg.PersistContext
	if pc == nil {
		pc = func() context.Context { return context.Background() }
	}

	e := &Engine{
		deps: evalDeps{
			Repo:  cfg.Repo,
			Helix: cfg.Helix,
		},
		persist:      pc,
		obs:          cfg.Obs,
		notify:       cfg.Notify,
		send:         cfg.Send,
		cooldown:     newCooldownTracker(),
		work:         make(chan workItem, workQueueSize),
		intervalNext: make(map[int64]time.Time),
	}

	e.rules.Store([]entity.Rule(nil))

	return e
}

// Start workers and interval ticker.
func (e *Engine) Start(ctx context.Context) {
	e.ctx, e.cancel = context.WithCancel(ctx)

	for i := 0; i < defaultWorkerCount; i++ {
		e.wg.Add(1)

		go e.workerLoop()
	}

	e.wg.Add(1)

	go e.scheduleLoop()
}

// Stop shuts down workers.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}

	e.wg.Wait()
}

// Reload replaces the in-memory rule snapshot (call after DB changes).
func (e *Engine) Reload(ctx context.Context, list []entity.Rule) {
	e.rules.Store(list)

	e.intervalMu.Lock()
	e.intervalNext = make(map[int64]time.Time)
	e.intervalMu.Unlock()
}

// HandleChatMessage dispatches chat_message rules.
func (e *Engine) HandleChatMessage(channel, user, text string) {
	p := EvalPayload{
		Event:    EventChatMessage,
		Channel:  trimLower(channel),
		Username: trimLower(user),
		Text:     text,
	}

	e.dispatchEvent(EventChatMessage, p)
}

// HandleStreamStart dispatches stream_start rules.
func (e *Engine) HandleStreamStart(channel, title string) {
	p := EvalPayload{
		Event:   EventStreamStart,
		Channel: trimLower(channel),
		Title:   title,
	}

	e.dispatchEvent(EventStreamStart, p)
}

// HandleStreamEnd dispatches stream_end rules.
func (e *Engine) HandleStreamEnd(channel string) {
	p := EvalPayload{
		Event:   EventStreamEnd,
		Channel: trimLower(channel),
	}

	e.dispatchEvent(EventStreamEnd, p)
}

// KeywordMatchChat returns true if any enabled chat_message rule passes all middlewares except cooldown is skipped.
func (e *Engine) KeywordMatchChat(ctx context.Context, channel, user, text string) bool {
	p := EvalPayload{
		Event:    EventChatMessage,
		Channel:  trimLower(channel),
		Username: trimLower(user),
		Text:     text,
	}

	for _, r := range e.snapshot() {
		if !r.Enabled || r.EventType != EventChatMessage {
			continue
		}

		if e.runChain(ctx, r, p, true) {
			return true
		}
	}

	return false
}
