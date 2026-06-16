// Package hub fans server-generated events out to all connected WebSocket
// clients, and remembers the current cycle's state so that a client connecting
// mid-debate can be brought up to speed with a single snapshot.
//
// Concurrency model: a single goroutine (Run) owns all mutable state — the set
// of clients and the cached snapshot state. Every interaction (a client
// connecting, a client leaving, the engine publishing an event) is delivered to
// that goroutine over a channel. Because only one goroutine ever touches the
// state, we need no mutexes and have no data races by construction.
package hub

import (
	"context"
	"log"
	"time"

	"aivaldebatten/internal/protocol"
)

// clientSendBuffer is how many envelopes we will queue for a client before
// declaring it too slow and dropping it. Large enough to absorb a burst (like
// the initial snapshot plus a few live events), small enough to notice a truly
// stuck client quickly.
const clientSendBuffer = 64

// Client represents one connected browser. The WebSocket handler creates it via
// Hub.Connect, then ranges over Out() to write frames to the socket.
type Client struct {
	send chan protocol.Envelope
}

// Out is the stream of envelopes the WebSocket handler should write to the
// socket. It is closed when the client is dropped or disconnected, which the
// handler can use as its signal to tear down the connection.
func (c *Client) Out() <-chan protocol.Envelope { return c.send }

// snapshotState is the hub's running picture of the current cycle, used to build
// a snapshot for late joiners. It is only ever mutated by the Run goroutine.
type snapshotState struct {
	cycleID     string
	topic       string
	endsAt      time.Time
	bots        []protocol.Bot
	messages    []protocol.Message
	typingBotID string
}

// Hub coordinates clients and event distribution.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	publish    chan command

	// Owned exclusively by Run:
	clients map[*Client]bool
	state   snapshotState
}

// New creates a Hub. Call Run (typically in its own goroutine) to start it.
func New() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		publish:    make(chan command),
		clients:    make(map[*Client]bool),
	}
}

// Run is the hub's single owning goroutine. It blocks until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Close all client channels so their writers shut down.
			for c := range h.clients {
				close(c.send)
				delete(h.clients, c)
			}
			return

		case c := <-h.register:
			// Send the new client a snapshot of the current state first, then
			// add it to the broadcast set so it receives subsequent events.
			if env, ok := h.buildSnapshot(); ok {
				c.send <- env // buffered; safe because buffer >= 1
			}
			h.clients[c] = true

		case c := <-h.unregister:
			if h.clients[c] {
				delete(h.clients, c)
				close(c.send)
			}

		case cmd := <-h.publish:
			// Update the cached state, then fan the event out to everyone.
			env, ok := cmd.apply(&h.state)
			if !ok {
				continue
			}
			h.broadcast(env)
		}
	}
}

// broadcast delivers env to every client, dropping any client whose buffer is
// full (i.e. it cannot keep up). Dropping closes its channel, signalling the
// writer goroutine to disconnect.
func (h *Hub) broadcast(env protocol.Envelope) {
	for c := range h.clients {
		select {
		case c.send <- env:
		default:
			log.Printf("hub: dropping slow client")
			delete(h.clients, c)
			close(c.send)
		}
	}
}

// buildSnapshot constructs a snapshot envelope from the current cached state.
// Returns ok=false if there is no active cycle yet (nothing to show).
func (h *Hub) buildSnapshot() (protocol.Envelope, bool) {
	if h.state.cycleID == "" {
		return protocol.Envelope{}, false
	}
	// Copy the messages slice so the marshaled snapshot is independent of future
	// mutation. (In practice marshaling happens here in the Run goroutine, but
	// copying keeps the intent explicit and safe.)
	msgs := make([]protocol.Message, len(h.state.messages))
	copy(msgs, h.state.messages)

	env, err := protocol.NewSnapshot(protocol.SnapshotData{
		CycleID:     h.state.cycleID,
		Topic:       h.state.topic,
		EndsAt:      h.state.endsAt,
		Bots:        h.state.bots,
		Messages:    msgs,
		TypingBotID: h.state.typingBotID,
	})
	if err != nil {
		log.Printf("hub: building snapshot: %v", err)
		return protocol.Envelope{}, false
	}
	return env, true
}

// ---------------------------------------------------------------------------
// Public API used by the WebSocket handler and the engine
// ---------------------------------------------------------------------------

// Connect registers a new client and returns it. The first thing the client
// will receive on Out() is a snapshot (if a cycle is in progress).
func (h *Hub) Connect() *Client {
	c := &Client{send: make(chan protocol.Envelope, clientSendBuffer)}
	h.register <- c
	return c
}

// Disconnect removes a client (e.g. the socket closed). Safe to call once.
func (h *Hub) Disconnect(c *Client) {
	h.unregister <- c
}

// PublishTopicChanged announces a new cycle.
func (h *Hub) PublishTopicChanged(d protocol.TopicChangedData) {
	h.publish <- topicChangedCmd(d)
}

// PublishTyping announces that a bot has begun typing.
func (h *Hub) PublishTyping(d protocol.BotTypingData) {
	h.publish <- typingCmd(d)
}

// PublishMessage delivers a completed message.
func (h *Hub) PublishMessage(d protocol.MessageData) {
	h.publish <- messageCmd(d)
}
