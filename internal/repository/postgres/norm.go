package postgres

import "strings"

func normalizeStoredUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
