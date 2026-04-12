package live

import "strings"

// NormalizeTwitchChannel returns a canonical lowercase channel name without #.
func NormalizeTwitchChannel(ch string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ch)), "#")
}

func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}

	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}
