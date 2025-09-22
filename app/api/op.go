package api

type IdMessage interface {
	GetId() string
}

func (m *WsMessageMessage) GetId() string {
	return m.Id
}

func (m *WsMessage) GetId() string {
	return m.Id
}
