package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

// ExpandTemplate replaces $VAR placeholders (uppercase ASCII key after $).
func ExpandTemplate(tpl string, vars map[string]string) string {
	if tpl == "" || len(vars) == 0 {
		return tpl
	}

	out := tpl

	for k, v := range vars {
		key := "$" + strings.ToUpper(k)
		out = strings.ReplaceAll(out, key, v)
	}

	return out
}

// TemplateVars builds standard variables for actions.
func TemplateVars(ruleID int64, channel, username, text, title string) map[string]string {
	return map[string]string{
		"RULE_ID":  fmt.Sprint(ruleID),
		"CHANNEL":  channel,
		"USERNAME": username,
		"TEXT":     text,
		"TITLE":    title,
	}
}

type evalDeps struct {
	Repo  repository.Store
	Helix *helix.Client
}

func (d *evalDeps) channelOnline(ctx context.Context, channelLogin string) (bool, error) {
	ch := trimLower(channelLogin)
	if ch == "" {
		return false, nil
	}

	bid, ok, err := d.Repo.MonitoredChannelTwitchUserID(ctx, ch)
	if err != nil {
		return false, err
	}

	if !ok {
		resolved, err := d.Helix.ResolveChannel(ctx, ch)
		if err != nil {
			return false, err
		}

		bid = resolved.ID
	}

	liveMap, err := d.Helix.HelixStreamsLiveByBroadcasterIDs(ctx, []int64{bid})
	if err != nil {
		return false, err
	}

	return liveMap[bid], nil
}

// MiddlewareOK runs one middleware; cooldown is handled in engine with lastFired map.
func MiddlewareOK(ctx context.Context, d *evalDeps, mw entity.RuleMiddleware, p EvalPayload, skipCooldown bool) bool {
	if skipCooldown && mw.Type == MWCooldown {
		return true
	}

	switch mw.Type {
	case MWFilterChannel:
		return mwFilterChannel(ctx, d, mw.Settings, p)
	case MWFilterUser:
		return mwFilterUser(mw.Settings, p)
	case MWMatchRegex:
		return mwMatchRegex(mw.Settings, p.Text)
	case MWContainsWord:
		return mwContainsWord(mw.Settings, p.Text)
	case MWCooldown:
		// evaluated in engine with mutex
		return true
	default:
		return false
	}
}

func mwFilterChannel(ctx context.Context, d *evalDeps, s map[string]any, p EvalPayload) bool {
	ch := trimLower(p.Channel)
	if ch == "" {
		return false
	}

	inc := strSliceFromAny(s["include_logins"])
	exc := strSliceFromAny(s["exclude_logins"])

	if len(inc) > 0 {
		found := false

		for _, x := range inc {
			if x == ch {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	for _, x := range exc {
		if x == ch {
			return false
		}
	}

	req, _ := s["require_online"].(bool)
	if !req {
		return true
	}

	live, err := d.channelOnline(ctx, ch)
	if err != nil {
		return false
	}

	return live
}

func mwFilterUser(s map[string]any, p EvalPayload) bool {
	u := trimLower(p.Username)

	inc := strSliceFromAny(s["include_logins"])
	exc := strSliceFromAny(s["exclude_logins"])

	if len(inc) > 0 && u == "" {
		return false
	}

	if len(inc) > 0 {
		found := false

		for _, x := range inc {
			if x == u {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	for _, x := range exc {
		if x == u {
			return false
		}
	}

	return true
}

func mwMatchRegex(s map[string]any, text string) bool {
	pat, _ := s["pattern"].(string)
	if pat == "" {
		return false
	}

	ci, _ := s["case_insensitive"].(bool)

	re, err := regexpCompile(pat, ci)
	if err != nil {
		return false
	}

	msg := text
	if r := []rune(msg); len(r) > maxRegexRunes {
		msg = string(r[:maxRegexRunes])
	}

	return re.MatchString(msg)
}

func regexpCompile(pattern string, caseInsensitive bool) (*regexp.Regexp, error) {
	if caseInsensitive {
		return regexp.Compile("(?i)" + pattern)
	}

	return regexp.Compile(pattern)
}

func mwContainsWord(s map[string]any, text string) bool {
	wordsAny, ok := s["words"].([]any)
	if !ok || len(wordsAny) == 0 {
		return false
	}

	lower := strings.ToLower(text)

	for _, w := range wordsAny {
		sw, ok := w.(string)
		if !ok {
			continue
		}

		sw = strings.ToLower(strings.TrimSpace(sw))
		if sw == "" {
			continue
		}

		if strings.Contains(lower, sw) {
			return true
		}
	}

	return false
}

type cooldownTracker struct {
	mu    sync.Mutex
	lastF map[int64]time.Time
}

func newCooldownTracker() *cooldownTracker {
	return &cooldownTracker{lastF: make(map[int64]time.Time)}
}

func (c *cooldownTracker) ok(ruleID int64, seconds int, now time.Time) bool {
	if seconds <= 0 {
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	prev, ok := c.lastF[ruleID]
	if ok && now.Sub(prev) < time.Duration(seconds)*time.Second {
		return false
	}

	return true
}

func (c *cooldownTracker) mark(ruleID int64, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastF[ruleID] = now
}
