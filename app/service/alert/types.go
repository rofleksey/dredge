package alert

import "regexp"

type Selector struct {
	Channel  *regexp.Regexp
	Username *regexp.Regexp
	Message  *regexp.Regexp
}
