package domain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultNotebook   = "Inbox"
	UntitledNoteTitle = "Untitled Note"
)

type Notebook string

type Note struct {
	Tags      []string
	Notebook  string
	UpdatedAt time.Time
	Favorite  bool
	Title     string
	File      string
	Format    string
}

func DefaultNote() Note {
	return Note{
		Tags:      []string{},
		Notebook:  DefaultNotebook,
		UpdatedAt: time.Now(),
		Title:     UntitledNoteTitle,
		File:      UntitledNoteTitle,
		Format:    "",
	}
}

func (n Note) Normalized() Note {
	if n.Tags == nil {
		n.Tags = []string{}
	}
	if strings.TrimSpace(n.Notebook) == "" {
		n.Notebook = DefaultNotebook
	}
	if strings.TrimSpace(n.Title) == "" {
		n.Title = UntitledNoteTitle
	}
	if strings.TrimSpace(n.File) == "" {
		n.File = fmt.Sprintf("%s.%s", n.Title, n.Format)
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = time.Now()
	}
	return n
}

func (n Note) String() string {
	n = n.Normalized()
	return fmt.Sprintf("%s/%s", n.Notebook, n.Title)
}

func (n Note) LegacyPath() string {
	n = n.Normalized()
	return n.File
}

func (n Note) Path() string {
	n = n.Normalized()
	return filepath.Join(n.Notebook, n.File)
}

func (n Note) MarshalJSON() ([]byte, error) {
	n = n.Normalized()
	type rawNote struct {
		Tags      []string  `json:"tags"`
		Notebook  string    `json:"notebook"`
		UpdatedAt time.Time `json:"updated_at"`
		Favorite  bool      `json:"favorite"`
		Title     string    `json:"title"`
		File      string    `json:"file"`
		Format    string    `json:"format"`
	}
	return json.Marshal(rawNote{
		Tags:      n.Tags,
		Notebook:  n.Notebook,
		UpdatedAt: n.UpdatedAt,
		Favorite:  n.Favorite,
		Title:     n.Title,
		File:      n.File,
		Format:    n.Format,
	})
}

func (n *Note) UnmarshalJSON(data []byte) error {
	type rawNote struct {
		Tags      []string  `json:"tags"`
		Notebook  string    `json:"notebook"`
		Folder    string    `json:"folder"`
		UpdatedAt time.Time `json:"updated_at"`
		Date      time.Time `json:"date"`
		Favorite  bool      `json:"favorite"`
		Title     string    `json:"title"`
		File      string    `json:"file"`
		Format    string    `json:"format"`
		Language  string    `json:"language"`
	}
	var raw rawNote
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	n.Tags = raw.Tags
	n.Notebook = raw.Notebook
	if n.Notebook == "" {
		n.Notebook = raw.Folder
	}
	n.UpdatedAt = raw.UpdatedAt
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = raw.Date
	}
	n.Favorite = raw.Favorite
	n.Title = raw.Title
	n.File = raw.File
	n.Format = raw.Format
	if n.Format == "" {
		n.Format = raw.Language
	}
	*n = n.Normalized()
	return nil
}

type Notes struct {
	Items []Note
}

func (n Notes) String(i int) string {
	return n.Items[i].String()
}

func (n Notes) Len() int {
	return len(n.Items)
}
