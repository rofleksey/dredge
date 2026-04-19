package twitch

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ValidateSPAReturnURL checks that candidate is a safe redirect target (same origin as configured oauth_return_url).
func (o *OAuth) ValidateSPAReturnURL(candidate string) error {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return nil
	}

	if len(candidate) > 2048 {
		return errors.New("return url too long")
	}

	base, err := url.Parse(o.returnURL)
	if err != nil {
		return fmt.Errorf("oauth: invalid configured return url: %w", err)
	}

	cand, err := url.Parse(candidate)
	if err != nil {
		return fmt.Errorf("invalid return url: %w", err)
	}

	if cand.Scheme != base.Scheme || cand.Host != base.Host {
		return errors.New("return url origin mismatch")
	}

	return nil
}
