package notes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Store defines the storage operations for Notes.
type Store interface {
	List(ctx context.Context, opts ListOptions) ([]Note, error)
	Get(ctx context.Context, id string) (Note, error)
	Save(ctx context.Context, note Note) (Note, error)
	Archive(ctx context.Context, id string, archived bool) error
	MoveToTrash(ctx context.Context, id string) error
}

// ListOptions specifies filters for listing notes.
type ListOptions struct {
	Query           string
	IncludeArchived bool
}

// FileStore implements the Store interface using local markdown files.
type FileStore struct {
	NotesDir string
	TrashDir string
}

// NewFileStore creates and initializes a new FileStore, creating directories if needed.
func NewFileStore(notesDir, trashDir string) (*FileStore, error) {
	expandedNotes, err := expandPath(notesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand notes path: %w", err)
	}

	expandedTrash, err := expandPath(trashDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand trash path: %w", err)
	}

	// Create directories if they do not exist
	if err := os.MkdirAll(expandedNotes, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory %s: %w", expandedNotes, err)
	}
	if err := os.MkdirAll(expandedTrash, 0755); err != nil {
		return nil, fmt.Errorf("failed to create trash directory %s: %w", expandedTrash, err)
	}

	return &FileStore{
		NotesDir: expandedNotes,
		TrashDir: expandedTrash,
	}, nil
}

// expandPath replaces ~ with the user's home directory and cleans the path.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Clean(path), nil
}

// List scans the notes directory and returns filtered, sorted notes.
func (f *FileStore) List(ctx context.Context, opts ListOptions) ([]Note, error) {
	entries, err := os.ReadDir(f.NotesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	var notes []Note
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}

		filePath := filepath.Join(f.NotesDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that cannot be read
			continue
		}

		note, parseErr := ParseNote(string(content), filePath)
		if parseErr != nil {
			// If malformed, ParseNote still returns a note structure with fallback title
			// and content as body. Ensure path is correct and keep it.
			note.Path = filePath
		}

		// Ensure ID is populated from filename if missing in frontmatter
		if note.ID == "" {
			note.ID = strings.TrimSuffix(entry.Name(), ".md")
		}

		notes = append(notes, note)
	}

	// Filter notes
	filtered := FilterNotes(notes, opts)

	// Sort notes by Position ascending, then UpdatedAt descending
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Position != filtered[j].Position {
			return filtered[i].Position < filtered[j].Position
		}
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})

	return filtered, nil
}

// Get loads a single note by its ID.
func (f *FileStore) Get(ctx context.Context, id string) (Note, error) {
	// First check direct file match
	directPath := filepath.Join(f.NotesDir, id+".md")
	content, err := os.ReadFile(directPath)
	if err == nil {
		note, err := ParseNote(string(content), directPath)
		if err != nil {
			return note, err
		}
		if note.ID == "" {
			note.ID = id
		}
		return note, nil
	}

	// Fallback scan if filename and ID don't match exactly
	entries, err := os.ReadDir(f.NotesDir)
	if err != nil {
		return Note{}, fmt.Errorf("failed to read notes directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		filePath := filepath.Join(f.NotesDir, entry.Name())
		c, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		note, err := ParseNote(string(c), filePath)
		if err == nil && note.ID == id {
			return note, nil
		}
	}

	return Note{}, os.ErrNotExist
}

// Save saves a note to disk.
func (f *FileStore) Save(ctx context.Context, note Note) (Note, error) {
	now := time.Now()
	note.Title = note.DisplayTitle()
	note.Tags = nil

	// Check if this is a new note
	isNew := note.ID == ""

	var originalPath string
	if !isNew {
		originalPath = note.Path
		if originalPath == "" {
			originalPath = filepath.Join(f.NotesDir, note.ID+".md")
		}
	}

	if isNew {
		// New note: generate timestamped ID
		note.CreatedAt = now
		note.UpdatedAt = now
		timestamp := now.Format("20060102-150405")
		slug := Slugify(note.DisplayTitle())
		note.ID = fmt.Sprintf("%s-%s", timestamp, slug)

		// Handle duplicate slugs by appending counter
		counter := 1
		for {
			targetPath := filepath.Join(f.NotesDir, note.ID+".md")
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				break
			}
			note.ID = fmt.Sprintf("%s-%s-%d", timestamp, slug, counter)
			counter++
		}
	} else {
		// Existing note: update timestamps without rewriting the file identity.
		note.UpdatedAt = now
		if note.CreatedAt.IsZero() {
			note.CreatedAt = now
		}
	}

	// Format note markdown
	content, err := FormatNote(note)
	if err != nil {
		return note, fmt.Errorf("failed to format note: %w", err)
	}

	// Rename file if existing note path changed
	newPath := filepath.Join(f.NotesDir, note.ID+".md")
	if originalPath != "" && originalPath != newPath {
		if err := os.Rename(originalPath, newPath); err != nil && !os.IsNotExist(err) {
			return note, fmt.Errorf("failed to rename file from %s to %s: %w", originalPath, newPath, err)
		}
	}

	note.Path = newPath

	// Write note content to disk
	if err := os.WriteFile(newPath, []byte(content), 0644); err != nil {
		return note, fmt.Errorf("failed to write note file: %w", err)
	}

	return note, nil
}

// Archive updates the archived state of a note.
func (f *FileStore) Archive(ctx context.Context, id string, archived bool) error {
	note, err := f.Get(ctx, id)
	if err != nil {
		return err
	}
	note.Archived = archived
	_, err = f.Save(ctx, note)
	return err
}

// MoveToTrash moves the note file to the trash directory.
func (f *FileStore) MoveToTrash(ctx context.Context, id string) error {
	note, err := f.Get(ctx, id)
	if err != nil {
		return err
	}

	if note.Path == "" {
		note.Path = filepath.Join(f.NotesDir, id+".md")
	}

	trashPath := filepath.Join(f.TrashDir, id+".md")

	// Handle duplicate filename in trash by appending a timestamp
	if _, err := os.Stat(trashPath); err == nil {
		timestamp := time.Now().Format("20060102-150405")
		trashPath = filepath.Join(f.TrashDir, fmt.Sprintf("%s-%s.md", id, timestamp))
	}

	if err := os.Rename(note.Path, trashPath); err != nil {
		return fmt.Errorf("failed to move file to trash: %w", err)
	}

	return nil
}

// Slugify converts a title into a clean terminal/file-friendly lowercase slug.
func Slugify(title string) string {
	title = strings.ToLower(title)
	var sb strings.Builder
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('-')
		}
	}
	s := sb.String()

	// Collapse multiple dashes
	reg := regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	if s == "" {
		return "untitled"
	}
	return s
}
