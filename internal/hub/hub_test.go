package hub

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"aivaldebatten/internal/protocol"
)

// recvWithin reads one envelope from a client or fails the test on timeout.
func recvWithin(t *testing.T, c *Client, d time.Duration) protocol.Envelope {
	t.Helper()
	select {
	case env, ok := <-c.Out():
		if !ok {
			t.Fatal("client channel closed unexpectedly")
		}
		return env
	case <-time.After(d):
		t.Fatal("timed out waiting for envelope")
		return protocol.Envelope{}
	}
}

func TestBroadcastAndLateJoinerSnapshot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := New()
	go h.Run(ctx)

	// Early client connects before any cycle: should get no snapshot yet.
	early := h.Connect()

	// Start a cycle.
	h.PublishTopicChanged(protocol.TopicChangedData{
		CycleID: "c1",
		Topic:   "Test topic",
		EndsAt:  time.Now().Add(time.Minute),
		Bots:    []protocol.Bot{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}},
	})

	// Early client should receive the topic_changed event.
	if env := recvWithin(t, early, time.Second); env.Type != protocol.EventTopicChanged {
		t.Fatalf("early client: expected topic_changed, got %s", env.Type)
	}

	// A bot speaks.
	msg := protocol.Message{ID: "m1", BotID: "a", Text: "hello", TS: time.Now()}
	h.PublishMessage(protocol.MessageData{Message: msg})
	if env := recvWithin(t, early, time.Second); env.Type != protocol.EventMessage {
		t.Fatalf("early client: expected message, got %s", env.Type)
	}

	// Late client joins mid-cycle: its FIRST event must be a snapshot that
	// already contains the earlier message and the current topic.
	late := h.Connect()
	env := recvWithin(t, late, time.Second)
	if env.Type != protocol.EventSnapshot {
		t.Fatalf("late client: expected snapshot first, got %s", env.Type)
	}
	var snap protocol.SnapshotData
	if err := json.Unmarshal(env.Data, &snap); err != nil {
		t.Fatalf("decoding snapshot: %v", err)
	}
	if snap.Topic != "Test topic" {
		t.Errorf("snapshot topic = %q, want %q", snap.Topic, "Test topic")
	}
	if len(snap.Messages) != 1 || snap.Messages[0].ID != "m1" {
		t.Errorf("snapshot should contain the prior message, got %+v", snap.Messages)
	}
}
