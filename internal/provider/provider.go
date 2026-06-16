// Package provider defines how the conversation engine obtains bot responses.
//
// This is the seam between "orchestration" (whose turn is it, what's the topic,
// fan messages out to clients) which lives in our backend, and "prompting"
// (actually producing the words a bot says) which you wanted to keep external.
//
// Today the only implementation is MockProvider. Later, an LLMProvider that
// calls an external API can implement the same interface and drop straight in
// with no change to the engine.
package provider

import (
	"context"
	"time"

	"aivaldebatten/internal/protocol"
)

// Request carries everything a provider needs to produce the next utterance.
type Request struct {
	// Topic is the current debate subject.
	Topic string
	// Speaker is the bot whose turn it is to talk (includes its Persona).
	Speaker protocol.Bot
	// Roster is the full set of participants, so a provider can refer to others
	// by name (e.g. rebut the previous speaker).
	Roster []protocol.Bot
	// History is the conversation so far, oldest first. May be empty at the
	// start of a cycle (the opening statement).
	History []protocol.Message
}

// Response is a produced utterance plus how long the bot should appear to be
// "typing" before the message lands. ThinkDelay is what makes the UI feel like
// a real chat rather than messages teleporting in.
type Response struct {
	Text       string
	ThinkDelay time.Duration
}

// Provider produces the next message for a given speaker. Implementations must
// respect ctx cancellation: if ctx is done (cycle reset or shutdown) they
// should return ctx.Err() promptly instead of blocking.
type Provider interface {
	Generate(ctx context.Context, req Request) (Response, error)
}
