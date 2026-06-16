package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"aivaldebatten/internal/config"
	"aivaldebatten/internal/protocol"
	"aivaldebatten/internal/provider"
)

// fakePublisher records the sequence of published event types under a mutex,
// since the engine publishes from its own goroutine.
type fakePublisher struct {
	mu    sync.Mutex
	types []protocol.EventType
}

func (f *fakePublisher) record(t protocol.EventType) {
	f.mu.Lock()
	f.types = append(f.types, t)
	f.mu.Unlock()
}

func (f *fakePublisher) PublishTopicChanged(protocol.TopicChangedData) {
	f.record(protocol.EventTopicChanged)
}
func (f *fakePublisher) PublishTyping(protocol.BotTypingData) { f.record(protocol.EventBotTyping) }
func (f *fakePublisher) PublishMessage(protocol.MessageData)  { f.record(protocol.EventMessage) }

func (f *fakePublisher) snapshot() []protocol.EventType {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]protocol.EventType, len(f.types))
	copy(out, f.types)
	return out
}

// fastProvider returns immediately with a tiny think delay so the test runs
// quickly and deterministically.
type fastProvider struct{}

func (fastProvider) Generate(ctx context.Context, req provider.Request) (provider.Response, error) {
	return provider.Response{Text: "stance", ThinkDelay: time.Millisecond}, nil
}

func TestRunCycleProducesTypingThenMessage(t *testing.T) {
	cfg := &config.Config{
		CycleDuration: config.Duration(120 * time.Millisecond),
		Topics:        []string{"Topic"},
		Bots:          []protocol.Bot{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}},
	}
	pub := &fakePublisher{}
	e := New(cfg, fastProvider{}, pub)
	e.interTurnPause = time.Millisecond // keep the test snappy

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	e.Run(ctx)

	types := pub.snapshot()
	if len(types) == 0 || types[0] != protocol.EventTopicChanged {
		t.Fatalf("expected first event to be topic_changed, got %v", types)
	}

	// Find the first typing event and assert it is followed by a message.
	foundPair := false
	for i := 0; i+1 < len(types); i++ {
		if types[i] == protocol.EventBotTyping && types[i+1] == protocol.EventMessage {
			foundPair = true
			break
		}
	}
	if !foundPair {
		t.Errorf("expected a typing->message pair in %v", types)
	}
}
