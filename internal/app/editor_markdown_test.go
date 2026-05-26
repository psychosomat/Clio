package app

import "testing"

func TestInsertSnippetIntoEmptyValue(t *testing.T) {
	updated, line, col := insertSnippet("", 0, 0, "\n## "+snippetCursor+"\n")

	want := "\n## \n"
	if updated != want {
		t.Fatalf("expected updated value %q, got %q", want, updated)
	}
	if line != 1 || col != 3 {
		t.Fatalf("expected cursor at line 1 col 3, got line %d col %d", line, col)
	}
}

func TestInsertSnippetAtCursor(t *testing.T) {
	value := "Hello world"
	updated, line, col := insertSnippet(value, 0, 5, "["+snippetCursor+"](https://)")

	want := "Hello[](https://) world"
	if updated != want {
		t.Fatalf("expected updated value %q, got %q", want, updated)
	}
	if line != 0 || col != 6 {
		t.Fatalf("expected cursor at line 0 col 6, got line %d col %d", line, col)
	}
}

func TestEditorCursorIndexClampsColumn(t *testing.T) {
	value := "abc\ndef"
	got := editorCursorIndex(value, 1, 99)

	if got != len([]rune("abc\ndef")) {
		t.Fatalf("expected cursor index at end of value, got %d", got)
	}
}
