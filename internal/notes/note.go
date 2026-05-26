package notes

import "time"

// Note represents a markdown note with its front matter and path.
type Note struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title,omitempty"`
	Body      string    `yaml:"-"` // Body is stored in the markdown body, not in the front matter YAML.
	Tags      []string  `yaml:"tags,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Position  int       `yaml:"position,omitempty"`
	Archived  bool      `yaml:"archived"`
	Folder    string    `yaml:"folder,omitempty"`
	Path      string    `yaml:"-"` // File path is not stored in YAML front matter.
}

// DisplayTitle derives a list title from the note body, falling back to legacy title metadata.
func (n Note) DisplayTitle() string {
	for _, line := range splitLines(n.Body) {
		title := normalizeDisplayLine(line)
		if title != "" {
			return title
		}
	}

	if normalizeDisplayLine(n.Title) != "" {
		return normalizeDisplayLine(n.Title)
	}

	return "Untitled"
}
