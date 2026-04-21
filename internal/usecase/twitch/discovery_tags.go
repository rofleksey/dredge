package twitch

import "strings"

// streamTagsCoverRequired returns true when required is empty, or when every non-empty required tag
// is present on the stream (case-insensitive, trimmed), AND semantics.
func streamTagsCoverRequired(required, streamTags []string) bool {
	if len(required) == 0 {
		return true
	}

	set := make(map[string]struct{}, len(streamTags))

	for _, t := range streamTags {
		k := strings.ToLower(strings.TrimSpace(t))
		if k == "" {
			continue
		}

		set[k] = struct{}{}
	}

	for _, r := range required {
		k := strings.ToLower(strings.TrimSpace(r))
		if k == "" {
			continue
		}

		if _, ok := set[k]; !ok {
			return false
		}
	}

	return true
}
