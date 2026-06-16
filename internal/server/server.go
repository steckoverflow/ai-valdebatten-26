// Package server wires HTTP routing: the WebSocket endpoint that streams the
// live debate, a health check, and the embedded SPA.
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"aivaldebatten/internal/hub"
	"aivaldebatten/internal/web"
)

const (
	// writeTimeout bounds a single frame write so one stuck client can't block
	// forever (belt-and-suspenders alongside the hub's slow-client dropping).
	writeTimeout = 10 * time.Second
	// pingInterval keeps idle connections and intermediaries alive and lets us
	// notice dead peers.
	pingInterval = 30 * time.Second
)

// Handler builds the top-level HTTP handler.
func Handler(h *hub.Hub) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler(h))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", web.Handler())
	return mux
}

// wsHandler upgrades the request to a WebSocket and streams hub events to the
// client until it disconnects.
func wsHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// The stream is read-only public data with no cookies or auth, so
			// cross-origin WebSocket hijacking is not a concern. Allowing all
			// origins also lets the dev Vite proxy connect without friction.
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Printf("ws: accept failed: %v", err)
			return
		}
		defer conn.CloseNow()

		// We never expect inbound messages. CloseRead drains/handles control
		// frames and gives us a context that cancels when the client leaves.
		ctx := conn.CloseRead(r.Context())

		client := h.Connect()
		defer h.Disconnect(client)

		ping := time.NewTicker(pingInterval)
		defer ping.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case env, ok := <-client.Out():
				if !ok {
					// Hub dropped us (e.g. too slow) and closed the channel.
					conn.Close(websocket.StatusPolicyViolation, "too slow")
					return
				}
				data, err := json.Marshal(env)
				if err != nil {
					log.Printf("ws: marshal: %v", err)
					continue
				}
				wctx, cancel := context.WithTimeout(ctx, writeTimeout)
				err = conn.Write(wctx, websocket.MessageText, data)
				cancel()
				if err != nil {
					return
				}

			case <-ping.C:
				pctx, cancel := context.WithTimeout(ctx, writeTimeout)
				err := conn.Ping(pctx)
				cancel()
				if err != nil {
					return
				}
			}
		}
	}
}
