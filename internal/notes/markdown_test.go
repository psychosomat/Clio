package notes

import (
	"strings"
	"testing"
)

func TestParseNote(t *testing.T) {
	content := `---
id: 20260526-001122-example-title
archived: false
---

Markdown body here.`

	note, err := ParseNote(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note.ID != "20260526-001122-example-title" {
		t.Errorf("expected ID 20260526-001122-example-title, got %s", note.ID)
	}
	if note.DisplayTitle() != "Markdown body here." {
		t.Errorf("expected display title from body, got %s", note.DisplayTitle())
	}
	if note.Archived {
		t.Errorf("expected archived to be false")
	}
	if strings.TrimSpace(note.Body) != "Markdown body here." {
		t.Errorf("expected Body 'Markdown body here.', got %q", note.Body)
	}
}

func TestParseNoteFallback(t *testing.T) {
	content := "Just raw text without front matter."
	note, err := ParseNote(content, "test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note.DisplayTitle() != content {
		t.Errorf("expected display title to use raw content, got %q", note.DisplayTitle())
	}
	if strings.TrimSpace(note.Body) != content {
		t.Errorf("expected Body to equal content, got %q", note.Body)
	}
}

func TestParseNoteMalformed(t *testing.T) {
	content := `---
id: test
title: malformed
malformed YAML [
`
	_, err := ParseNote(content, "test.md")
	if err == nil {
		t.Error("expected error due to malformed yaml")
	}
}

func TestFormatNote(t *testing.T) {
	note := Note{
		ID:   "test-id",
		Body: "Test body contents.",
	}

	formatted, err := FormatNote(note)
	if err != nil {
		t.Fatalf("unexpected format error: %v", err)
	}

	if strings.Contains(formatted, "title:") {
		t.Error("formatted output should not persist title metadata")
	}
	if strings.Contains(formatted, "tags:") {
		t.Error("formatted output should not persist tags metadata")
	}
	if !strings.Contains(formatted, "Test body contents.") {
		t.Error("formatted output missing body")
	}
}

func TestExtractTags(t *testing.T) {
	body := `
# Header
This is a #test of the tag #extraction system.
Let's make sure #work-item and #idea_1 are captured.
But we should not match #123 (numeric) or colors like #fff or #AABBCC.
Also, # Header2 is a header, not a tag.
`
	tags := ExtractTags(body)
	expected := []string{"test", "extraction", "work-item", "idea_1"}

	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %v", len(expected), tags)
	}

	for i, exp := range expected {
		if tags[i] != exp {
			t.Errorf("expected tag %d to be %s, got %s", i, exp, tags[i])
		}
	}
}

func TestMergeTags(t *testing.T) {
	front := []string{"work", "idea"}
	body := []string{"idea", "project", "WORK"}
	merged := MergeTags(front, body)

	expected := []string{"work", "idea", "project"}
	if len(merged) != len(expected) {
		t.Fatalf("expected %d merged tags, got %v", len(expected), merged)
	}
	for i, exp := range expected {
		if merged[i] != exp {
			t.Errorf("expected %s, got %s", exp, merged[i])
		}
	}
}
