package messages

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

func parseSearchQuery(query string) (*SearchParams, error) {
	var params SearchParams

	usernamePattern := regexp.MustCompile(`username:([^\s]+)`)
	channelPattern := regexp.MustCompile(`channel:([^\s]+)`)
	datePattern := regexp.MustCompile(`date:([^~\s]+)~([^\s]+)`)

	if matches := usernamePattern.FindStringSubmatch(query); matches != nil {
		params.Username = matches[1]
		query = strings.Replace(query, matches[0], "", 1)
	}

	if matches := channelPattern.FindStringSubmatch(query); matches != nil {
		params.Channel = matches[1]
		query = strings.Replace(query, matches[0], "", 1)
	}

	if matches := datePattern.FindStringSubmatch(query); matches != nil {
		fromStr, toStr := matches[1], matches[2]

		if fromTime, err := parseDate(fromStr); err == nil {
			params.DateFrom = &fromTime
		}

		if toTime, err := parseDate(toStr); err == nil {
			params.DateTo = &toTime
		}

		query = strings.Replace(query, matches[0], "", 1)
	}

	params.TextQuery = strings.TrimSpace(query)

	return &params, nil
}

func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"01/02/2006",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func buildSearchQuery(params *SearchParams, offset, limit int) (string, []interface{}, error) {
	query := sq.Select("id", "created", "username", "channel", "text").
		From("messages").
		OrderBy("created DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if params.Username != "" {
		query = query.Where(sq.Eq{"username": params.Username})
	}

	if params.Channel != "" {
		query = query.Where(sq.Eq{"channel": params.Channel})
	}

	if params.DateFrom != nil {
		query = query.Where(sq.GtOrEq{"created": *params.DateFrom})
	}

	if params.DateTo != nil {
		query = query.Where(sq.LtOrEq{"created": *params.DateTo})
	}

	if params.TextQuery != "" {
		query = query.Where(sq.ILike{"text": "%" + params.TextQuery + "%"})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, args, nil
}
