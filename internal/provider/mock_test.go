package provider

import (
	"context"
	"strings"
	"testing"

	"aivaldebatten/internal/protocol"
)

func TestMockProviderUsesSpeakerManifesto(t *testing.T) {
	m := NewMock()
	req := Request{
		Topic: "kollektivtrafik",
		Speaker: protocol.Bot{
			ID:        "parti",
			Name:      "Partiet",
			Manifesto: "Vår unika linje ska höras. En annan möjlig formulering.",
		},
	}

	resp, err := m.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if !strings.Contains(resp.Text, "Vår unika linje ska höras.") && !strings.Contains(resp.Text, "En annan möjlig formulering.") {
		t.Fatalf("response did not include manifesto text: %q", resp.Text)
	}
}
