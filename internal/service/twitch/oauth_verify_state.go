package twitch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// VerifyState checks the HMAC and expiry on a state value from Twitch's redirect.
// It returns an optional SPA return URL embedded in state (empty = use configured oauth_return_url).
func (o *OAuth) VerifyState(state string) (spaReturnURL string, err error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid state")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid state payload: %w", err)
	}

	wantSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid state sig: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(o.hmacSecret))
	_, _ = mac.Write(payload)
	got := mac.Sum(nil)

	if !hmac.Equal(wantSig, got) {
		return "", errors.New("state signature mismatch")
	}

	var p oauthStatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return "", fmt.Errorf("invalid state json: %w", err)
	}

	if time.Now().Unix() > p.Exp {
		return "", errors.New("state expired")
	}

	if p.Ret != "" {
		if err := o.ValidateSPAReturnURL(p.Ret); err != nil {
			return "", fmt.Errorf("state return url: %w", err)
		}
	}

	o.stateMu.Lock()
	defer o.stateMu.Unlock()

	if o.usedStateNonces == nil {
		o.usedStateNonces = make(map[string]int64)
	}

	now := time.Now().Unix()

	for n, exp := range o.usedStateNonces {
		if now > exp {
			delete(o.usedStateNonces, n)
		}
	}

	if _, dup := o.usedStateNonces[p.N]; dup {
		return "", errors.New("state already used")
	}

	o.usedStateNonces[p.N] = p.Exp

	return p.Ret, nil
}
