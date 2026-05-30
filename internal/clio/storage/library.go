package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	appconfig "clio/internal/clio/config"
	"clio/internal/clio/domain"
)

type Repository struct {
	config appconfig.App
}

func New(config appconfig.App) *Repository {
	return &Repository{config: config}
}

func (r *Repository) ReadAll() ([]domain.Note, bool, error) {
	if err := os.MkdirAll(r.config.Home, 0o755); err != nil {
		return nil, false, err
	}
	path := r.metadataPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, false, err
		}
		legacyData, legacyErr := os.ReadFile(r.legacyMetadataPath())
		if legacyErr != nil {
			if !errors.Is(legacyErr, fs.ErrNotExist) {
				return nil, false, legacyErr
			}
			if writeErr := os.WriteFile(path, []byte("[]"), 0o644); writeErr != nil {
				return nil, false, writeErr
			}
			return []domain.Note{}, false, nil
		}
		notes, _, err := unmarshalNotes(legacyData)
		return notes, err == nil, err
	}
	return unmarshalNotes(data)
}

func (r *Repository) WriteAll(notes []domain.Note) error {
	if err := os.MkdirAll(r.config.Home, 0o755); err != nil {
		return err
	}
	payload, err := json.Marshal(sortNotes(notes))
	if err != nil {
		return err
	}
	return os.WriteFile(r.metadataPath(), payload, 0o644)
}

func (r *Repository) Scan(notes []domain.Note) ([]domain.Note, bool, error) {
	if err := os.MkdirAll(r.config.Home, 0o755); err != nil {
		return notes, false, err
	}
	modified := false
	exists := map[string]struct{}{}
	for _, note := range notes {
		exists[note.Path()] = struct{}{}
	}

	homeEntries, err := os.ReadDir(r.config.Home)
	if err != nil {
		return notes, false, err
	}

	for _, homeEntry := range homeEntries {
		if !homeEntry.IsDir() {
			continue
		}
		if strings.HasPrefix(homeEntry.Name(), ".") {
			continue
		}
		notebookPath := filepath.Join(r.config.Home, homeEntry.Name())
		notebookEntries, err := os.ReadDir(notebookPath)
		if err != nil {
			continue
		}
		for _, notebookEntry := range notebookEntries {
			if notebookEntry.IsDir() {
				continue
			}
			notePath := filepath.Join(homeEntry.Name(), notebookEntry.Name())
			if _, ok := exists[notePath]; ok {
				continue
			}
			ext := filepath.Ext(notebookEntry.Name())
			info, err := notebookEntry.Info()
			updatedAt := time.Now()
			if err == nil {
				updatedAt = info.ModTime()
			}
			notes = append(notes, domain.Note{
				Notebook:  homeEntry.Name(),
				UpdatedAt: updatedAt,
				Title:     strings.TrimSuffix(notebookEntry.Name(), ext),
				File:      notebookEntry.Name(),
				Format:    strings.TrimPrefix(ext, "."),
				Tags:      []string{},
			}.Normalized())
			modified = true
		}
	}

	kept := make([]domain.Note, 0, len(notes))
	for _, note := range notes {
		if _, err := os.Stat(r.Path(note)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				modified = true
				continue
			}
			return notes, modified, err
		}
		kept = append(kept, note.Normalized())
	}
	return sortNotes(kept), modified, nil
}

func (r *Repository) Create(note domain.Note) error {
	note = note.Normalized()
	if err := os.MkdirAll(filepath.Dir(r.Path(note)), 0o755); err != nil {
		return err
	}
	file, err := os.Create(r.Path(note))
	if err != nil {
		return err
	}
	return file.Close()
}

func (r *Repository) WriteContent(note domain.Note, content string) error {
	note = note.Normalized()
	if err := os.MkdirAll(filepath.Dir(r.Path(note)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(r.Path(note), []byte(content), 0o644)
}

func (r *Repository) AppendContent(note domain.Note, content string) error {
	note = note.Normalized()
	if err := os.MkdirAll(filepath.Dir(r.Path(note)), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(r.Path(note), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func (r *Repository) ReadContent(note domain.Note) (string, error) {
	content, err := os.ReadFile(r.Path(note.Normalized()))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (r *Repository) Delete(note domain.Note) error {
	return os.Remove(r.Path(note.Normalized()))
}

func (r *Repository) Rename(previous, next domain.Note) error {
	previous = previous.Normalized()
	next = next.Normalized()
	oldPath := r.Path(previous)
	newPath := r.Path(next)
	if oldPath == newPath {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (r *Repository) CreateFolder(name string) error {
	dir := filepath.Join(r.config.Home, name)
	return os.MkdirAll(dir, 0o755)
}

func (r *Repository) DeleteFolder(name string) error {
	dir := filepath.Join(r.config.Home, name)
	return os.RemoveAll(dir)
}

func (r *Repository) Path(note domain.Note) string {
	return filepath.Join(r.config.Home, note.Normalized().Path())
}

func (r *Repository) metadataPath() string {
	return filepath.Join(r.config.Home, r.config.File)
}

func (r *Repository) legacyMetadataPath() string {
	return filepath.Join(r.config.Home, "snippets.json")
}

func unmarshalNotes(data []byte) ([]domain.Note, bool, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return []domain.Note{}, false, nil
	}
	var notes []domain.Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, false, err
	}
	return sortNotes(notes), true, nil
}

func sortNotes(notes []domain.Note) []domain.Note {
	normalized := make([]domain.Note, 0, len(notes))
	for _, note := range notes {
		normalized = append(normalized, note.Normalized())
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].UpdatedAt.Equal(normalized[j].UpdatedAt) {
			return normalized[i].String() < normalized[j].String()
		}
		return normalized[i].UpdatedAt.After(normalized[j].UpdatedAt)
	})
	return normalized
}
