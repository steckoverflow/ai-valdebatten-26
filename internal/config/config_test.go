package config

import (
	"testing"

	"aivaldebatten/internal/protocol"
)

// TestLoadSampleConfig loads the real config/config.json shipped with the repo
// and asserts the basics. It doubles as a guard against JSON typos in the
// committed config file.
func TestLoadSampleConfig(t *testing.T) {
	// Tests run with the package directory as the working dir, so walk up to
	// the repo root to find the config file.
	cfg, err := Load("../../config/config.json")
	if err != nil {
		t.Fatalf("loading sample config: %v", err)
	}
	if cfg.CycleDuration.Std() <= 0 {
		t.Errorf("expected positive cycle duration, got %s", cfg.CycleDuration.Std())
	}
	if len(cfg.Topics) == 0 {
		t.Errorf("expected at least one topic")
	}
	if len(cfg.Bots) < 2 {
		t.Errorf("expected at least two bots, got %d", len(cfg.Bots))
	}
}

// TestValidateRejectsBadConfig confirms validation catches common mistakes.
func TestValidateRejectsBadConfig(t *testing.T) {
	good := []protocol.Bot{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
	}
	dup := []protocol.Bot{
		{ID: "a", Name: "A"},
		{ID: "a", Name: "B"},
	}
	min := Duration(1)

	cases := map[string]Config{
		"no topics":     {CycleDuration: min, Bots: good},
		"one bot":       {CycleDuration: min, Topics: []string{"x"}, Bots: good[:1]},
		"zero duration": {Topics: []string{"x"}, Bots: good},
		"duplicate ids": {CycleDuration: min, Topics: []string{"x"}, Bots: dup},
	}
	for name, c := range cases {
		if err := c.validate(); err == nil {
			t.Errorf("%s: expected validation error, got nil", name)
		}
	}
}
