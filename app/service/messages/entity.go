package messages

import "time"

type SearchParams struct {
	Username  string
	Channel   string
	DateFrom  *time.Time
	DateTo    *time.Time
	TextQuery string
}
