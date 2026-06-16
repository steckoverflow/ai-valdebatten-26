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
// with a sentence from the speaker's configured platform, and fakes a human-like
// typing delay.
//
// It implements provider.Provider.
type MockProvider struct{}

// NewMock returns a ready-to-use MockProvider.
func NewMock() *MockProvider { return &MockProvider{} }

// openers are used for the first message of a cycle (no prior history).
var openers = []string{
	"När det gäller %q är min utgångspunkt enkel.",
	"Låt mig vara tydlig om %q.",
	"I frågan om %q har vi inte råd att tveka.",
	"Så här ser jag på %q.",
}

// rebuttals reference the previous speaker by name (%s is their name).
var rebuttals = []string{
	"Jag håller inte med %s här.",
	"%s har en poäng, men missar något viktigt.",
	"Med respekt för %s håller inte det argumentet hela vägen.",
	"Om vi bygger vidare på det %s sa:",
	"%s har delvis rätt, men också delvis fel.",
}

// fallbackStances are used only when a bot has no configured manifesto/persona.
var fallbackStances = []string{
	"Det viktigaste är att politiken fungerar i människors vardag.",
	"Vi måste väga ambitioner mot vad samhället faktiskt klarar av.",
	"Det här handlar om prioriteringar och ansvar för framtiden.",
	"En ansvarsfull politik måste hålla även efter nästa val.",
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
	b.WriteString(stanceFor(req))

	text := b.String()

	return Response{
		Text:       text,
		ThinkDelay: typingDelay(text),
	}, nil
}

// stanceFor lets the configured bot platform shape mock responses. This keeps
// the demo aligned with the roster without hard-coding party IDs here.
func stanceFor(req Request) string {
	if candidates := sentenceCandidates(req.Speaker.Manifesto); len(candidates) > 0 {
		return pick(candidates)
	}
	if candidates := sentenceCandidates(req.Speaker.Persona); len(candidates) > 0 {
		return pick(candidates)
	}
	return pick(fallbackStances)
}

func sentenceCandidates(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	parts := strings.Split(text, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part+".")
	}
	return out
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
