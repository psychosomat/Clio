package notes

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStore(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "clio-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	notesDir := filepath.Join(tempDir, "notes")
	trashDir := filepath.Join(tempDir, "trash")

	store, err := NewFileStore(notesDir, trashDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}

	ctx := context.Background()

	// 1. Verify directories are created
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		t.Error("expected notes directory to be created")
	}
	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		t.Error("expected trash directory to be created")
	}

	// 2. Test saving a new note
	note1 := Note{
		Body: "This is my #first note! #golang",
	}

	savedNote1, err := store.Save(ctx, note1)
	if err != nil {
		t.Fatalf("failed to save note: %v", err)
	}

	if savedNote1.ID == "" {
		t.Error("expected ID to be generated")
	}
	if !strings.HasSuffix(savedNote1.ID, "-this-is-my-first-note-golang") {
		t.Errorf("expected ID to end with body-derived slug, got %s", savedNote1.ID)
	}
	if savedNote1.Title != "This is my #first note! #golang" {
		t.Errorf("expected display title derived from body, got %q", savedNote1.Title)
	}
	if len(savedNote1.Tags) != 0 {
		t.Errorf("expected tags to be removed from saved note metadata, got %v", savedNote1.Tags)
	}

	// 3. Test list notes
	list, err := store.List(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("failed to list notes: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 note, got %d", len(list))
	}
	if list[0].ID != savedNote1.ID {
		t.Errorf("expected note ID %s, got %s", savedNote1.ID, list[0].ID)
	}

	// 4. Test updating note body keeps a stable ID
	savedNote1.Body = "Hello Universe"
	updatedNote1, err := store.Save(ctx, savedNote1)
	if err != nil {
		t.Fatalf("failed to update note: %v", err)
	}

	if updatedNote1.ID != savedNote1.ID {
		t.Errorf("expected existing note ID to stay stable, got %s -> %s", savedNote1.ID, updatedNote1.ID)
	}

	if _, err := os.Stat(updatedNote1.Path); os.IsNotExist(err) {
		t.Errorf("expected new file %s to exist", updatedNote1.Path)
	}

	// 5. Test archiving a note
	err = store.Archive(ctx, updatedNote1.ID, true)
	if err != nil {
		t.Fatalf("failed to archive note: %v", err)
	}

	// Active list should now be empty
	list, err = store.List(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected active list to be empty, got %d", len(list))
	}

	// List with IncludeArchived should show the note
	list, err = store.List(ctx, ListOptions{IncludeArchived: true})
	if err != nil {
		t.Fatalf("failed to list archived: %v", err)
	}
	if len(list) != 1 || !list[0].Archived {
		t.Errorf("expected 1 archived note, got %v", list)
	}

	// Unarchive
	err = store.Archive(ctx, updatedNote1.ID, false)
	if err != nil {
		t.Fatalf("failed to unarchive: %v", err)
	}

	// 6. Test duplicate slugs during creation
	noteDup1 := Note{Body: "Duplicate Title"}
	noteDup2 := Note{Body: "Duplicate Title"}

	savedDup1, err := store.Save(ctx, noteDup1)
	if err != nil {
		t.Fatalf("failed to save duplicate 1: %v", err)
	}
	// Sleep for a fraction to ensure different timestamp if needed, but since our store uses seconds,
	// if we save instantly, it might use the exact same second. We want to test that if it does, it resolves it!
	// Let's force duplicate ID generation path by modifying the mock note ID or using identical timestamp string.
	// Since our duplicate resolver checks file existence in NotesDir, let's write a file to force collision.
	savedDup2, err := store.Save(ctx, noteDup2)
	if err != nil {
		t.Fatalf("failed to save duplicate 2: %v", err)
	}

	if savedDup1.ID == savedDup2.ID {
		t.Errorf("expected unique IDs, both were %s", savedDup1.ID)
	}

	// 7. Test moving to trash (delete scenario)
	err = store.MoveToTrash(ctx, updatedNote1.ID)
	if err != nil {
		t.Fatalf("failed to delete note: %v", err)
	}

	// Check notes list is 2 (only the duplicates)
	list, err = store.List(ctx, ListOptions{IncludeArchived: true})
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 notes remaining, got %d", len(list))
	}

	// Check trash contains the file
	trashFiles, err := os.ReadDir(trashDir)
	if err != nil {
		t.Fatalf("failed to read trash: %v", err)
	}
	if len(trashFiles) != 1 {
		t.Errorf("expected 1 file in trash, got %d", len(trashFiles))
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Simple Title", "simple-title"},
		{"Title with !@# Special Characters!", "title-with-special-characters"},
		{"   leading/trailing spaces   ", "leading-trailing-spaces"},
		{"multiple--dashes__and_spaces", "multiple-dashes-and-spaces"},
		{"", "untitled"},
		{"!!!", "untitled"},
	}

	for _, tc := range tests {
		got := Slugify(tc.title)
		if got != tc.expected {
			t.Errorf("Slugify(%q) = %q; expected %q", tc.title, got, tc.expected)
		}
	}
}
