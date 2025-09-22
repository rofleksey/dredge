package mapper

import (
	"dredge/app/api"
	"dredge/pkg/database"
)

func MapMessage(m database.Message) api.Message {
	return api.Message{
		Channel:  m.Channel,
		Id:       m.ID,
		Text:     m.Text,
		Username: m.Username,
		Created:  m.Created,
	}
}
