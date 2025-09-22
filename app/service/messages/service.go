package messages

import (
	"context"
	"dredge/pkg/database"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
)

type Service struct {
	dbConn  *pgxpool.Pool
	queries *database.Queries
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		dbConn:  do.MustInvoke[*pgxpool.Pool](di),
		queries: do.MustInvoke[*database.Queries](di),
	}, nil
}

// Search parses and executes search queries with the format:
// "username:<username> channel:<channel> date:<from>~<to> <text substring>"
func (s *Service) Search(ctx context.Context, query string, offset int, limit int) ([]database.Message, error) {
	params, err := parseSearchQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	sqlQuery, args, err := buildSearchQuery(params, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}

	rows, err := s.dbConn.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var messages []database.Message
	for rows.Next() {
		var msg database.Message

		err = rows.Scan(&msg.ID, &msg.Created, &msg.Username, &msg.Channel, &msg.Text)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}

func (s *Service) TotalCount(ctx context.Context) (int64, error) {
	count, err := s.queries.CountMessages(ctx)
	if err != nil {
		return 0, fmt.Errorf("CountMessages: %w", err)
	}

	return count, err
}
