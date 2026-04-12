package live

import (
	"regexp"
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

// Limits regex evaluation input size to reduce ReDoS risk from admin-defined patterns.
const maxRegexMessageRunes = 4000

type compiledRule struct {
	entity   entity.Rule
	re       *regexp.Regexp
	incUsers map[string]struct{} // empty means use * (all)
	allUsers bool
	denUsers map[string]struct{}
	incChans map[string]struct{}
	allChans bool
	denChans map[string]struct{}
}

func parseLoginSet(s string, starMeansAll bool) (all bool, set map[string]struct{}) {
	s = strings.TrimSpace(s)
	if starMeansAll && s == "*" {
		return true, nil
	}

	if !starMeansAll && s == "" {
		return true, nil
	}

	set = make(map[string]struct{})

	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}

		set[p] = struct{}{}
	}

	return len(set) == 0 && starMeansAll, set
}

func compileRules(rules []entity.Rule) ([]compiledRule, []error) {
	var (
		out  []compiledRule
		errs []error
	)

	for _, r := range rules {
		re, err := regexp.Compile(r.Regex)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		allU, incU := parseLoginSet(r.IncludedUsers, true)
		_, denU := parseLoginSet(r.DeniedUsers, false)
		allC, incC := parseLoginSet(r.IncludedChannels, true)
		_, denC := parseLoginSet(r.DeniedChannels, false)
		out = append(out, compiledRule{
			entity:   r,
			re:       re,
			incUsers: incU,
			allUsers: allU,
			denUsers: denU,
			incChans: incC,
			allChans: allC,
			denChans: denC,
		})
	}

	return out, errs
}

func (c *compiledRule) matches(chatterLogin, channelLogin, message string) bool {
	ch := strings.ToLower(strings.TrimSpace(chatterLogin))
	chChan := strings.ToLower(strings.TrimSpace(channelLogin))

	if !c.allUsers {
		if _, ok := c.incUsers[ch]; !ok {
			return false
		}
	}

	if len(c.denUsers) > 0 {
		if _, ok := c.denUsers[ch]; ok {
			return false
		}
	}

	if !c.allChans {
		if _, ok := c.incChans[chChan]; !ok {
			return false
		}
	}

	if len(c.denChans) > 0 {
		if _, ok := c.denChans[chChan]; ok {
			return false
		}
	}

	msg := message
	if r := []rune(msg); len(r) > maxRegexMessageRunes {
		msg = string(r[:maxRegexMessageRunes])
	}

	return c.re.MatchString(msg)
}
