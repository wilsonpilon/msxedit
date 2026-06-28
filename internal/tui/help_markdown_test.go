package tui

import "testing"

func TestNewHelpContentLoadsMarkdownFile(t *testing.T) {
	hc := NewHelpContent()
	if hc == nil {
		t.Fatal("expected help content")
	}
	if hc.currentTopicID != "contents" {
		t.Fatalf("expected root topic contents, got %q", hc.currentTopicID)
	}
	for _, id := range []string{"contents", "editor_commands", "block_commands", "syntax_highlighting"} {
		if _, ok := hc.topics[id]; !ok {
			t.Fatalf("expected topic %q to exist", id)
		}
	}
	if got := len(hc.topics["editor_commands"].Links); got < 5 {
		t.Fatalf("expected editor_commands to have at least 5 links, got %d", got)
	}
}

