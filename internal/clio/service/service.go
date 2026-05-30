package service

import (
	"bytes"
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	appconfig "clio/internal/clio/config"
	"clio/internal/clio/domain"
	"clio/internal/clio/storage"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/glamour"
	"github.com/sahilm/fuzzy"
)

type Service struct {
	Config     appconfig.App
	Repository *storage.Repository
	mdRenderer *glamour.TermRenderer
}

func New(config appconfig.App) *Service {
	return &Service{
		Config:     config,
		Repository: storage.New(config),
	}
}

func (s *Service) Bootstrap() ([]domain.Note, error) {
	notes, needsWrite, err := s.Repository.ReadAll()
	if err != nil {
		return nil, err
	}
	notes, scanned, err := s.Repository.Scan(notes)
	if err != nil {
		return nil, err
	}
	if needsWrite || scanned {
		if err := s.Repository.WriteAll(notes); err != nil {
			return nil, err
		}
	}
	return sortNotes(notes), nil
}

func (s *Service) Persist(notes []domain.Note) error {
	return s.Repository.WriteAll(sortNotes(notes))
}

func (s *Service) SaveNote(content string, args []string, notes []domain.Note) ([]domain.Note, error) {
	name := domain.UntitledNoteTitle
	if len(args) > 0 {
		name = strings.Join(args, " ")
	}
	notebook, title, format := ParseName(name)
	note := domain.Note{
		Notebook:  notebook,
		UpdatedAt: time.Now(),
		Title:     title,
		File:      fmt.Sprintf("%s.%s", title, format),
		Format:    format,
		Tags:      []string{},
	}.Normalized()
	if err := s.Repository.WriteContent(note, content); err != nil {
		return notes, err
	}
	notes = append([]domain.Note{note}, notes...)
	return sortNotes(notes), s.Persist(notes)
}

func (s *Service) FindNote(search string, notes []domain.Note) domain.Note {
	matches := fuzzy.FindFrom(search, domain.Notes{Items: notes})
	if len(matches) == 0 {
		return domain.Note{}
	}
	return notes[matches[0].Index]
}

func (s *Service) RenderContent(note domain.Note, highlight bool) string {
	content, err := s.Repository.ReadContent(note)
	if err != nil {
		return ""
	}
	if !highlight {
		return content
	}
	mdFormats := map[string]bool{"md": true, "markdown": true, "": true}
	if mdFormats[note.Format] {
		return s.RenderMarkdown(content)
	}
	var b bytes.Buffer
	if err := quick.Highlight(&b, content, note.Format, "terminal16m", s.Config.Theme); err != nil {
		return content
	}
	return b.String()
}

func (s *Service) RenderMarkdown(content string) string {
	if s.mdRenderer == nil {
		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dracula"),
			glamour.WithWordWrap(0),
		)
		if err != nil {
			return content
		}
		s.mdRenderer = r
	}
	out, err := s.mdRenderer.Render(content)
	if err != nil {
		return content
	}
	return out
}

func (s *Service) CreateNote(currentNotebook string) (domain.Note, error) {
	notebook := currentNotebook
	if strings.TrimSpace(notebook) == "" {
		notebook = domain.DefaultNotebook
	}
	note := domain.Note{
		Notebook:  notebook,
		UpdatedAt: time.Now(),
		Title:     domain.UntitledNoteTitle,
		File:      fmt.Sprintf("note-%d", rand.Intn(1000000)),
		Format:    "",
		Tags:      []string{},
	}.Normalized()
	return note, s.Repository.Create(note)
}

func (s *Service) UpdateNote(previous, next domain.Note) (domain.Note, error) {
	next = next.Normalized()
	next.UpdatedAt = time.Now()
	next.File = fmt.Sprintf("%s.%s", next.Title, next.Format)
	if err := s.Repository.Rename(previous, next); err != nil {
		return previous, err
	}
	return next, nil
}

func (s *Service) AppendToNote(note domain.Note, content string) error {
	return s.Repository.AppendContent(note, content)
}

func (s *Service) DeleteNote(note domain.Note) error {
	return s.Repository.Delete(note)
}

func (s *Service) CreateFolder(name string) error {
	return s.Repository.CreateFolder(name)
}

func (s *Service) NotePath(note domain.Note) string {
	return s.Repository.Path(note)
}

func ParseName(value string) (string, string, string) {
	notebook := domain.DefaultNotebook
	title := domain.UntitledNoteTitle
	format := ""
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return notebook, title, format
	}
	if before, after, ok := strings.Cut(trimmed, "/"); ok {
		if before != "" {
			notebook = before
		}
		trimmed = after
	}
	ext := filepath.Ext(trimmed)
	if ext != "" {
		format = strings.TrimPrefix(ext, ".")
		trimmed = strings.TrimSuffix(trimmed, ext)
	}
	if strings.TrimSpace(trimmed) != "" {
		title = trimmed
	}
	return notebook, title, format
}

func sortNotes(notes []domain.Note) []domain.Note {
	sorted := make([]domain.Note, 0, len(notes))
	for _, note := range notes {
		sorted = append(sorted, note.Normalized())
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].UpdatedAt.Equal(sorted[j].UpdatedAt) {
			return sorted[i].String() < sorted[j].String()
		}
		return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
	})
	return sorted
}
