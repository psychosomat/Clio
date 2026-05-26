package app

import (
	"testing"

	"clio/internal/notes"
)

type stubOpener struct {
	targets []string
}

func (s *stubOpener) Open(target string) error {
	s.targets = append(s.targets, target)
	return nil
}

func TestPreviewLinkNavigationAndExternalOpen(t *testing.T) {
	store, err := notes.NewFileStore(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	config := ReadConfig()
	m := NewModel(store, config, nil)
	opener := &stubOpener{}
	m.Opener = opener
	m.Mode = editorMode
	m.width = 100
	m.height = 30
	m.EditorPreview = true
	m.EditorBody.SetValue("[one](https://example.com) [two](https://example.org)")

	m.updateMarkdownPreview()
	if len(m.PreviewLinks) != 2 {
		t.Fatalf("expected 2 preview links, got %d", len(m.PreviewLinks))
	}
	if m.PreviewLinkIndex != 0 {
		t.Fatalf("expected active preview link 0, got %d", m.PreviewLinkIndex)
	}

	m.cyclePreviewLink(1)
	m.updateMarkdownPreview()
	if m.PreviewLinkIndex != 1 {
		t.Fatalf("expected active preview link 1, got %d", m.PreviewLinkIndex)
	}

	cmd := m.openPreviewLink()
	if cmd == nil {
		t.Fatal("expected openPreviewLink to return command")
	}
	msg := cmd()
	if _, ok := msg.(MsgLinkOpened); !ok {
		t.Fatalf("expected MsgLinkOpened, got %T", msg)
	}
	if len(opener.targets) != 1 || opener.targets[0] != "https://example.org" {
		t.Fatalf("unexpected open targets: %+v", opener.targets)
	}
}

func TestPreviewWikiLinkOpensLocalNote(t *testing.T) {
	store, err := notes.NewFileStore(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	config := ReadConfig()
	m := NewModel(store, config, nil)
	m.width = 100
	m.height = 30
	m.Mode = editorMode
	m.EditorPreview = true
	m.Notes = []notes.Note{
		{ID: "note-1", Body: "# Section\n\nBody"},
	}
	m.EditorBody.SetValue("[[note-1#section|jump]]")

	m.updateMarkdownPreview()
	if len(m.PreviewLinks) != 1 {
		t.Fatalf("expected 1 preview link, got %d", len(m.PreviewLinks))
	}
}
