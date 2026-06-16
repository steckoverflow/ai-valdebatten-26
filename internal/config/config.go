// Package config loads and validates the runtime configuration: the bot
// roster, the list of debate topics, and tunable settings such as the cycle
// duration. It is loaded once at startup; if the file is missing or invalid we
// fail fast with a descriptive error rather than starting in a broken state.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"aivaldebatten/internal/protocol"
)

// Duration is a thin wrapper around time.Duration that can be unmarshaled from
// a human-friendly JSON string like "15m" or "90s". The standard library only
// understands an integer number of nanoseconds, which is awful to hand-write,
// so we parse it ourselves.
type Duration time.Duration

// UnmarshalJSON parses values like "15m" into a Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"15m\": %w", err)
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// Std returns the value as a standard library time.Duration for use elsewhere.
func (d Duration) Std() time.Duration { return time.Duration(d) }

// Config is the top-level configuration document.
type Config struct {
	// CycleDuration is how long each debate runs before a new topic is chosen.
	CycleDuration Duration `json:"cycleDuration"`
	// Topics is the pool of debate subjects; one is chosen at random per cycle.
	Topics []string `json:"topics"`
	// Bots is the participant roster.
	Bots []protocol.Bot `json:"bots"`
}

// Load reads and validates the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	// DisallowUnknownFields would be stricter, but we keep it lenient so adding
	// future fields doesn't break older binaries. Decode normally instead.
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config %q: %w", path, err)
	}
	return &cfg, nil
}

// validate enforces the minimum requirements for a sensible debate.
func (c *Config) validate() error {
	if c.CycleDuration.Std() <= 0 {
		return fmt.Errorf("cycleDuration must be positive (e.g. \"15m\")")
	}
	if len(c.Topics) == 0 {
		return fmt.Errorf("at least one topic is required")
	}
	// A debate needs at least two participants to take turns.
	if len(c.Bots) < 2 {
		return fmt.Errorf("at least two bots are required, got %d", len(c.Bots))
	}

	seen := make(map[string]bool, len(c.Bots))
	for i, b := range c.Bots {
		if b.ID == "" {
			return fmt.Errorf("bots[%d] is missing an id", i)
		}
		if seen[b.ID] {
			return fmt.Errorf("duplicate bot id %q", b.ID)
		}
		seen[b.ID] = true
		if b.Name == "" {
			return fmt.Errorf("bot %q is missing a name", b.ID)
		}
	}
	return nil
}
