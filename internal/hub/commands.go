package hub

import (
	"log"

	"aivaldebatten/internal/protocol"
)

// command is something the engine asks the hub to publish. Each command knows
// how to (a) update the hub's cached snapshot state and (b) produce the wire
// envelope to broadcast. Bundling both here keeps the Run loop trivial and
// ensures the cached state and the broadcast event can never drift apart.
type command interface {
	// apply mutates the cached state and returns the envelope to broadcast.
	// ok=false means "skip" (e.g. marshaling failed).
	apply(state *snapshotState) (protocol.Envelope, bool)
}

// --- topic_changed -------------------------------------------------------

type topicChangedCmd protocol.TopicChangedData

func (c topicChangedCmd) apply(state *snapshotState) (protocol.Envelope, bool) {
	// A new cycle resets the conversation history and typing indicator.
	state.cycleID = c.CycleID
	state.topic = c.Topic
	state.endsAt = c.EndsAt
	state.bots = c.Bots
	state.messages = nil
	state.typingBotID = ""

	env, err := protocol.NewTopicChanged(protocol.TopicChangedData(c))
	if err != nil {
		log.Printf("hub: marshaling topic_changed: %v", err)
		return protocol.Envelope{}, false
	}
	return env, true
}

// --- bot_typing ----------------------------------------------------------

type typingCmd protocol.BotTypingData

func (c typingCmd) apply(state *snapshotState) (protocol.Envelope, bool) {
	state.typingBotID = c.BotID

	env, err := protocol.NewBotTyping(protocol.BotTypingData(c))
	if err != nil {
		log.Printf("hub: marshaling bot_typing: %v", err)
		return protocol.Envelope{}, false
	}
	return env, true
}

// --- message -------------------------------------------------------------

type messageCmd protocol.MessageData

func (c messageCmd) apply(state *snapshotState) (protocol.Envelope, bool) {
	// A completed message clears the typing indicator and is appended to the
	// running history so late joiners see it in their snapshot.
	state.typingBotID = ""
	state.messages = append(state.messages, c.Message)

	env, err := protocol.NewMessage(protocol.MessageData(c))
	if err != nil {
		log.Printf("hub: marshaling message: %v", err)
		return protocol.Envelope{}, false
	}
	return env, true
}
