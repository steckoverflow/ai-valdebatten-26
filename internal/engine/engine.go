// Package engine runs the autonomous debate: it owns the 15-minute cycle, picks
// topics, decides whose turn it is, asks the provider for each utterance, and
// publishes events to the hub. It is the only place that knows the "rules" of
// the conversation; distribution to clients is the hub's job and producing the
// words is the provider's job.
package engine

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"aivaldebatten/internal/config"
	"aivaldebatten/internal/protocol"
	"aivaldebatten/internal/provider"
)

// Publisher is the subset of the hub the engine needs. Depending on an
// interface (rather than *hub.Hub) keeps the engine testable with a fake.
type Publisher interface {
	PublishTopicChanged(protocol.TopicChangedData)
	PublishTyping(protocol.BotTypingData)
	PublishMessage(protocol.MessageData)
}

// Engine drives the debate.
type Engine struct {
	cfg      *config.Config
	provider provider.Provider
	pub      Publisher

	// interTurnPause is a short "let readers catch up" gap inserted after each
	// message before the next bot starts typing.
	interTurnPause time.Duration
}

// New builds an Engine.
func New(cfg *config.Config, p provider.Provider, pub Publisher) *Engine {
	return &Engine{
		cfg:            cfg,
		provider:       p,
		pub:            pub,
		interTurnPause: 1500 * time.Millisecond,
	}
}

// Run executes cycles back-to-back until ctx is cancelled. Each cycle lasts
// cfg.CycleDuration and consists of bots taking turns to fill the whole window.
func (e *Engine) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		e.runCycle(ctx)
	}
}

// runCycle executes a single debate cycle: choose a topic, announce it, then
// loop turns until the cycle's time budget is spent (or the server shuts down).
func (e *Engine) runCycle(parent context.Context) {
	// Each cycle has its own deadline. When it fires, any in-flight provider
	// call and any typing delay abort, and we return to start the next cycle.
	cycleCtx, cancel := context.WithTimeout(parent, e.cfg.CycleDuration.Std())
	defer cancel()

	topic := e.cfg.Topics[rand.IntN(len(e.cfg.Topics))]
	cycleID := newCycleID()
	endsAt := time.Now().Add(e.cfg.CycleDuration.Std())

	log.Printf("engine: new cycle %s topic=%q", cycleID, topic)

	e.pub.PublishTopicChanged(protocol.TopicChangedData{
		CycleID: cycleID,
		Topic:   topic,
		EndsAt:  endsAt,
		Bots:    e.cfg.Bots,
	})

	var (
		history []protocol.Message
		lastBot string
		turnSeq int
	)

	for {
		if cycleCtx.Err() != nil {
			return // cycle window elapsed or shutting down
		}

		speaker := e.pickSpeaker(lastBot)

		// Show the typing indicator before we "think".
		e.pub.PublishTyping(protocol.BotTypingData{BotID: speaker.ID})

		resp, err := e.provider.Generate(cycleCtx, provider.Request{
			Topic:   topic,
			Speaker: speaker,
			Roster:  e.cfg.Bots,
			History: history,
		})
		if err != nil {
			// Almost always ctx cancellation (cycle ended). Stop the cycle.
			return
		}

		// Simulate the time spent typing the message.
		if !sleep(cycleCtx, resp.ThinkDelay) {
			return
		}

		msg := protocol.Message{
			ID:    fmt.Sprintf("%s-%d", cycleID, turnSeq),
			BotID: speaker.ID,
			Text:  resp.Text,
			TS:    time.Now(),
		}
		history = append(history, msg)
		lastBot = speaker.ID
		turnSeq++

		e.pub.PublishMessage(protocol.MessageData{Message: msg})

		// Brief pause so readers can keep up before the next bot starts.
		if !sleep(cycleCtx, e.interTurnPause) {
			return
		}
	}
}

// pickSpeaker chooses the next bot uniformly at random, avoiding an immediate
// repeat of the previous speaker (so no bot replies to itself). With only one
// bot this would loop forever, but config validation guarantees at least two.
func (e *Engine) pickSpeaker(lastBot string) protocol.Bot {
	for {
		b := e.cfg.Bots[rand.IntN(len(e.cfg.Bots))]
		if b.ID != lastBot {
			return b
		}
	}
}

// sleep waits for d or until ctx is cancelled. It returns true if the full
// duration elapsed, false if ctx was cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// newCycleID returns a short, reasonably-unique identifier for a cycle.
func newCycleID() string {
	return fmt.Sprintf("%d-%04x", time.Now().Unix(), rand.IntN(0x10000))
}
