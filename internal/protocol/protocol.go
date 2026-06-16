// Package protocol defines the shared domain types and the wire format used
// to communicate from the backend to connected clients over WebSocket.
//
// It is intentionally dependency-free (imports only the standard library) so
// that every other package in the project can import it without risking an
// import cycle. Think of it as the single source of truth for "what a Bot is,
// what a Message is, and what the server is allowed to send to the browser".
package protocol

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// Bot is one participant in the debate. The fields here are the ones we are
// willing to expose to the frontend (e.g. to render a name + colored bubble).
//
// Persona is the human-written description of the bot's personality/stance. It
// doubles as context for the response provider (the mock now, an LLM later).
//
// Manifesto is a longer-form statement of the bot's core beliefs/platform. Like
// Persona it can feed the provider as context, and the UI may surface it (e.g.
// in a profile/expandable bio).
type Bot struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Persona   string `json:"persona"`
	Manifesto string `json:"manifesto"`
	Color     string `json:"color"` // e.g. "#e63946", used by the UI for the bubble
}

// Message is a single utterance produced by a bot during a cycle.
type Message struct {
	ID    string    `json:"id"`
	BotID string    `json:"botId"`
	Text  string    `json:"text"`
	TS    time.Time `json:"ts"`
}

// ---------------------------------------------------------------------------
// Event type tags
// ---------------------------------------------------------------------------

// EventType is the discriminator carried in every Envelope so the client knows
// how to decode the accompanying Data payload.
type EventType string

const (
	// EventSnapshot is sent once, immediately after a client connects (or
	// reconnects). It contains the entire current state so a late joiner sees
	// the debate exactly as it stands right now.
	EventSnapshot EventType = "snapshot"

	// EventTopicChanged announces the start of a new 15-minute cycle: a fresh
	// topic, the bot roster, and when the cycle ends.
	EventTopicChanged EventType = "topic_changed"

	// EventBotTyping indicates a bot has started "typing" its next message.
	// The UI shows a typing indicator until the matching EventMessage arrives.
	EventBotTyping EventType = "bot_typing"

	// EventMessage delivers a completed bot message.
	EventMessage EventType = "message"
)

// ---------------------------------------------------------------------------
// Event payloads
// ---------------------------------------------------------------------------

// SnapshotData is the full current state, used for late joiners / reconnects.
// Replaying these fields lets a freshly connected client render an identical
// view to everyone else without waiting for the next live event.
type SnapshotData struct {
	CycleID     string    `json:"cycleId"`
	Topic       string    `json:"topic"`
	EndsAt      time.Time `json:"endsAt"`
	Bots        []Bot     `json:"bots"`
	Messages    []Message `json:"messages"`
	TypingBotID string    `json:"typingBotId,omitempty"` // empty when nobody is typing
}

// TopicChangedData announces a new cycle.
type TopicChangedData struct {
	CycleID string    `json:"cycleId"`
	Topic   string    `json:"topic"`
	EndsAt  time.Time `json:"endsAt"`
	Bots    []Bot     `json:"bots"`
}

// BotTypingData says which bot is currently composing a message.
type BotTypingData struct {
	BotID string `json:"botId"`
}

// MessageData wraps a completed message.
type MessageData struct {
	Message Message `json:"message"`
}

// ---------------------------------------------------------------------------
// Envelope
// ---------------------------------------------------------------------------

// Envelope is the outer frame for every message sent to clients. Type tells the
// client which concrete payload Data holds.
type Envelope struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

// newEnvelope marshals an arbitrary payload into an Envelope. It is unexported
// because callers should use the typed constructors below, which prevent
// pairing the wrong Type with the wrong payload.
func newEnvelope(t EventType, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: t, Data: raw}, nil
}

// The following constructors are the only intended way to build an Envelope.
// Keeping them here means the type tag and payload shape can never drift apart.

func NewSnapshot(d SnapshotData) (Envelope, error) {
	return newEnvelope(EventSnapshot, d)
}

func NewTopicChanged(d TopicChangedData) (Envelope, error) {
	return newEnvelope(EventTopicChanged, d)
}

func NewBotTyping(d BotTypingData) (Envelope, error) {
	return newEnvelope(EventBotTyping, d)
}

func NewMessage(d MessageData) (Envelope, error) {
	return newEnvelope(EventMessage, d)
}
