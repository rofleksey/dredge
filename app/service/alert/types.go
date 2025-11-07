package alert

import (
	"dredge/app/config"
	"regexp"
)

type Selector struct {
	AlertEntry config.AlertEntry
	MsgRegex   *regexp.Regexp
}
