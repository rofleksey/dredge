package rules

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (e *Engine) workerLoop() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case it, ok := <-e.work:
			if !ok {
				return
			}

			e.execute(it)
		}
	}
}

func (e *Engine) scheduleLoop() {
	defer e.wg.Done()

	t := time.NewTicker(schedulerTick)
	defer t.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case now := <-t.C:
			e.tickIntervals(now)
		}
	}
}

func (e *Engine) tickIntervals(now time.Time) {
	list := e.snapshot()

	e.intervalMu.Lock()
	defer e.intervalMu.Unlock()

	for _, r := range list {
		if !r.Enabled || r.EventType != EventInterval {
			continue
		}

		ch, _ := r.EventSettings["channel"].(string)
		ch = trimLower(ch)

		sec, ok := numFromMap(r.EventSettings, "interval_seconds")
		if !ok || sec <= 0 || ch == "" {
			continue
		}

		next, ok := e.intervalNext[r.ID]
		if !ok || next.IsZero() {
			e.intervalNext[r.ID] = now.Add(time.Duration(sec) * time.Second)

			continue
		}

		if now.Before(next) {
			continue
		}

		p := EvalPayload{
			Event:       EventInterval,
			Channel:     ch,
			IntervalSec: sec,
		}

		e.enqueueWork(r, p)

		e.intervalNext[r.ID] = now.Add(time.Duration(sec) * time.Second)
	}
}

func (e *Engine) snapshot() []entity.Rule {
	v := e.rules.Load()
	if v == nil {
		return nil
	}

	out, _ := v.([]entity.Rule)
	return out
}

func (e *Engine) dispatchEvent(event string, p EvalPayload) {
	for _, r := range e.snapshot() {
		if !r.Enabled || r.EventType != event {
			continue
		}

		if !e.ruleMatchesEventSettings(r, p) {
			continue
		}

		e.enqueueWork(r, p)
	}
}

func (e *Engine) ruleMatchesEventSettings(r entity.Rule, p EvalPayload) bool {
	if r.EventType != EventInterval {
		return true
	}

	want, _ := r.EventSettings["channel"].(string)
	want = trimLower(want)

	return want != "" && want == p.Channel
}

func (e *Engine) enqueueWork(r entity.Rule, p EvalPayload) {
	it := workItem{rule: r, payload: p}

	if r.UseSharedPool {
		select {
		case e.work <- it:
		default:
			go e.execute(it)
		}

		return
	}

	go e.execute(it)
}

func (e *Engine) execute(it workItem) {
	ctx := e.persist()
	if ctx == nil {
		ctx = context.Background()
	}

	if !e.runChain(ctx, it.rule, it.payload, false) {
		return
	}

	e.execAction(ctx, it.rule, it.payload)
}

func (e *Engine) runChain(ctx context.Context, rule entity.Rule, p EvalPayload, skipCooldown bool) bool {
	for _, mw := range rule.Middlewares {
		if mw.Type == MWCooldown {
			if skipCooldown {
				continue
			}

			secF, _ := numFromMap(mw.Settings, "seconds")
			sec := int(secF)
			if !e.cooldown.ok(rule.ID, sec, time.Now()) {
				return false
			}

			continue
		}

		if !MiddlewareOK(ctx, &e.deps, mw, p, skipCooldown) {
			return false
		}
	}

	return true
}

func (e *Engine) execAction(ctx context.Context, rule entity.Rule, p EvalPayload) {
	markCooldown := false

	for _, mw := range rule.Middlewares {
		if mw.Type == MWCooldown {
			markCooldown = true

			break
		}
	}

	switch rule.ActionType {
	case ActionNotify:
		tpl, _ := rule.ActionSettings["text"].(string)
		if tpl == "" && p.Event == EventInterval {
			tpl = "[interval] #$CHANNEL"
		}

		vars := TemplateVars(rule.ID, p.Channel, p.Username, p.Text, p.Title)
		out := ExpandTemplate(tpl, vars)

		switch p.Event {
		case EventChatMessage:
			e.notify.NotifyChatKeyword(ctx, p.Channel, p.Username, p.Text, out)
		case EventStreamStart:
			e.notify.NotifyStreamStart(ctx, p.Channel, p.Title, out)
		case EventStreamEnd:
			e.notify.NotifyStreamEnd(ctx, p.Channel, out)
		case EventInterval:
			e.notify.NotifyChatKeyword(ctx, p.Channel, "", "", out)
		default:
			e.notify.NotifyChatKeyword(ctx, p.Channel, p.Username, p.Text, out)
		}
	case ActionSendChat:
		aidF, _ := numFromMap(rule.ActionSettings, "account_id")
		aid := int64(aidF)
		chTpl, _ := rule.ActionSettings["channel"].(string)
		msgTpl, _ := rule.ActionSettings["message"].(string)

		vars := TemplateVars(rule.ID, p.Channel, p.Username, p.Text, p.Title)
		ch := ExpandTemplate(chTpl, vars)
		msg := ExpandTemplate(msgTpl, vars)

		err := e.send.SendMessage(ctx, aid, ch, msg)
		if err != nil {
			if e.obs != nil {
				e.obs.Logger.Debug("rules send_chat failed", zap.Error(err), zap.Int64("rule_id", rule.ID))
			}

			return
		}
	default:
		return
	}

	if markCooldown {
		e.cooldown.mark(rule.ID, time.Now())
	}
}
