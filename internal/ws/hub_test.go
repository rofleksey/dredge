package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	t.Parallel()

	h := NewHub("")
	require.NotNil(t, h)
}

func TestHub_BroadcastJSON_noClients(t *testing.T) {
	t.Parallel()

	h := NewHub("")
	h.BroadcastJSON(map[string]string{"k": "v"})
}

func TestHub_Upgrade_and_read(t *testing.T) {
	t.Parallel()

	h := NewHub("")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = h.Upgrade(w, r, 1)
	}))
	defer srv.Close()

	u := "ws" + strings.TrimPrefix(srv.URL, "http")

	d := websocket.Dialer{HandshakeTimeout: 2 * time.Second}

	conn, _, err := d.Dial(u, nil)
	require.NoError(t, err)

	require.NoError(t, conn.Close())

	time.Sleep(50 * time.Millisecond)

	h.BroadcastJSON(map[string]any{"type": "ping"})
}

func TestHub_Upgrade_initialPayloads(t *testing.T) {
	t.Parallel()

	h := NewHub("")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = h.Upgrade(w, r, 1, map[string]any{"type": "hello", "n": 1})
	}))
	defer srv.Close()

	u := "ws" + strings.TrimPrefix(srv.URL, "http")

	d := websocket.Dialer{HandshakeTimeout: 2 * time.Second}

	conn, _, err := d.Dial(u, nil)
	require.NoError(t, err)

	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	var got map[string]any

	require.NoError(t, conn.ReadJSON(&got))
	require.Equal(t, "hello", got["type"])
	require.EqualValues(t, 1, got["n"])
}
