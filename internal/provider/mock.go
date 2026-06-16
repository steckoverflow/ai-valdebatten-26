package provider

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// MockProvider fabricates plausible debate lines without any external service.
// It exists so we can build and demo the entire system before wiring up a real
// LLM. It is deliberately simple: it stitches a conversational opener together
// with a generic "stance" sentence, and fakes a human-like typing delay.
//
// It implements provider.Provider.
type MockProvider struct{}

// NewMock returns a ready-to-use MockProvider.
func NewMock() *MockProvider { return &MockProvider{} }

// openers are used for the first message of a cycle (no prior history).
var openers = []string{
	"On the question of %q, my position is simple.",
	"Let me be direct about %q.",
	"When it comes to %q, we cannot afford to hesitate.",
	"Here is how I see %q.",
}

// rebuttals reference the previous speaker by name (%s is their name).
var rebuttals = []string{
	"I have to disagree with %s here.",
	"%s makes a fair point, but it misses something.",
	"With respect to %s, that argument doesn't hold up.",
	"Building on what %s said —",
	"%s is half right, and half wrong.",
}

// stances are generic debate-flavored payload sentences.
var stances = []string{
	"The real cost always falls on ordinary people, and we ignore that at our peril.",
	"We have to weigh ambition against what taxpayers can actually bear.",
	"Future generations will judge us by whether we acted boldly today.",
	"Freedom means letting people decide for themselves, not mandating outcomes.",
	"The evidence from other countries is clear if we bother to look at it.",
	"This is about priorities, and right now ours are upside down.",
	"A responsible society plans for the long term, not the next election.",
	"Markets solve this better than any committee ever could.",
}

// Generate produces a fake response. It respects ctx cancellation so that a
// cycle reset or shutdown doesn't leave the engine waiting.
func (m *MockProvider) Generate(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}

	var b strings.Builder
	if len(req.History) == 0 {
		// Opening statement: reference the topic.
		fmt.Fprintf(&b, pick(openers), req.Topic)
	} else {
		// Rebuttal: reference whoever spoke last, by display name.
		prevName := m.nameOf(req, req.History[len(req.History)-1].BotID)
		fmt.Fprintf(&b, pick(rebuttals), prevName)
	}
	b.WriteString(" ")
	b.WriteString(pick(stances))

	text := b.String()

	return Response{
		Text:       text,
		ThinkDelay: typingDelay(text),
	}, nil
}

// nameOf resolves a bot ID to its display name using the roster, falling back
// to the raw ID if not found.
func (m *MockProvider) nameOf(req Request, id string) string {
	for _, bot := range req.Roster {
		if bot.ID == id {
			return bot.Name
		}
	}
	return id
}

// pick returns a random element from a non-empty slice.
func pick(s []string) string {
	return s[rand.IntN(len(s))]
}

// typingDelay fakes how long a human would take to type the message: a base
// "thinking" pause plus time proportional to length, with jitter, capped so a
// long message never stalls the debate for too long.
func typingDelay(text string) time.Duration {
	const (
		base    = 700 * time.Millisecond
		perRune = 18 * time.Millisecond
		max     = 5 * time.Second
	)
	d := base + time.Duration(len([]rune(text)))*perRune
	// Add +/-20% jitter so cadence feels natural.
	jitter := 0.8 + 0.4*rand.Float64()
	d = time.Duration(float64(d) * jitter)
	if d > max {
		d = max
	}
	return d
}
