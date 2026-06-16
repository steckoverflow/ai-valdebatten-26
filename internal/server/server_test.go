package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"aivaldebatten/internal/config"
	"aivaldebatten/internal/engine"
	"aivaldebatten/internal/hub"
	"aivaldebatten/internal/protocol"
	"aivaldebatten/internal/provider"
)

// TestEndToEndStream boots the full stack behind an httptest server and verifies
// a connecting client receives live debate events over the WebSocket.
func TestEndToEndStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{
		CycleDuration: config.Duration(10 * time.Second),
		Topics:        []string{"Integration topic"},
		Bots:          []protocol.Bot{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}},
	}

	h := hub.New()
	go h.Run(ctx)
	eng := engine.New(cfg, provider.NewMock(), h)
	go eng.Run(ctx)

	ts := httptest.NewServer(Handler(h))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	dialCtx, dialCancel := context.WithTimeout(ctx, 3*time.Second)
	defer dialCancel()
	conn, _, err := websocket.Dial(dialCtx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// The first frame will be a snapshot (the engine started the cycle before we
	// connected). We then expect the live typing->message rhythm. Collect event
	// types until we've seen both a bot_typing and a message, or time out.
	deadline := time.Now().Add(12 * time.Second)
	seen := map[protocol.EventType]bool{}
	for !(seen[protocol.EventBotTyping] && seen[protocol.EventMessage]) {
		if time.Now().After(deadline) {
			t.Fatalf("timed out; saw events: %v", seen)
		}
		readCtx, readCancel := context.WithTimeout(ctx, 6*time.Second)
		_, data, err := conn.Read(readCtx)
		readCancel()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var env protocol.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		seen[env.Type] = true
	}

	// Sanity: the very first frame for a mid-cycle joiner should be a snapshot.
	if !seen[protocol.EventSnapshot] {
		t.Errorf("expected an initial snapshot, saw: %v", seen)
	}
}

// TestServesIndex confirms the embedded SPA (placeholder or built) is served.
func TestServesIndex(t *testing.T) {
	h := hub.New()
	ts := httptest.NewServer(Handler(h))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("index status = %d, want 200", resp.StatusCode)
	}
}
