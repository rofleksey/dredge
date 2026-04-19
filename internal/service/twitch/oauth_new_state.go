package twitch

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// NewState builds a signed OAuth state parameter (CSRF protection). spaReturnURL is optional (empty = use configured return URL only).
func (o *OAuth) NewState(spaReturnURL string) (string, error) {
	if err := o.ValidateSPAReturnURL(spaReturnURL); err != nil {
		return "", err
	}

	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	p := oauthStatePayload{
		Exp: time.Now().Add(15 * time.Minute).Unix(),
		N:   hex.EncodeToString(raw[:]),
		Ret: strings.TrimSpace(spaReturnURL),
	}

	payload, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(o.hmacSecret))
	_, _ = mac.Write(payload)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}
