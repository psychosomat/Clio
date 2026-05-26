package notes

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var tagRegex = regexp.MustCompile(`\B#([a-zA-Z][a-zA-Z0-9_-]*)`)

type noteFrontMatter struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title,omitempty"`
	Tags      []string  `yaml:"tags,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Position  int       `yaml:"position,omitempty"`
	Archived  bool      `yaml:"archived"`
	Folder    string    `yaml:"folder,omitempty"`
}

// ParseNote parses a markdown note file's content into a Note struct.
// If the note has malformed front matter or YAML, it populates a warning title and returns the error
// rather than crashing, as per edge case requirements.
func ParseNote(content string, filePath string) (Note, error) {
	var note Note
	note.Path = filePath

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	if len(lines) == 0 || len(content) == 0 {
		note.Title = "Untitled"
		note.CreatedAt = time.Now()
		note.UpdatedAt = time.Now()
		return note, nil
	}

	// Check if we start with front matter delimiter
	firstLine := strings.TrimSpace(lines[0])
	if firstLine != "---" {
		// No front matter, treat the whole thing as body
		note.Title = "Untitled"
		note.Body = content
		note.CreatedAt = time.Now()
		note.UpdatedAt = time.Now()
		return note, nil
	}

	// Find the closing ---
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		// Missing closing divider
		note.Title = "Malformed Note"
		note.Body = content
		note.CreatedAt = time.Now()
		note.UpdatedAt = time.Now()
		return note, errors.New("malformed front matter: missing closing divider")
	}

	yamlContent := strings.Join(lines[1:closingIdx], "\n")
	var bodyContent string
	if closingIdx+1 < len(lines) {
		bodyContent = strings.Join(lines[closingIdx+1:], "\n")
	}

	err := yaml.Unmarshal([]byte(yamlContent), &note)
	if err != nil {
		note.Title = "Malformed YAML"
		note.Body = content
		note.CreatedAt = time.Now()
		note.UpdatedAt = time.Now()
		return note, fmt.Errorf("malformed front matter yaml: %w", err)
	}

	if note.Title == "" {
		note.Title = "Untitled"
	}
	note.Body = bodyContent

	// Standardize all parsed tags to lowercase
	for i, tag := range note.Tags {
		note.Tags[i] = strings.ToLower(strings.TrimSpace(tag))
	}

	return note, nil
}

// FormatNote serializes a Note struct into markdown content with front matter.
func FormatNote(note Note) (string, error) {
	fm := noteFrontMatter{
		ID:        note.ID,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
		Position:  note.Position,
		Archived:  note.Archived,
		Folder:    note.Folder,
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal front matter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n")

	body := strings.TrimSpace(note.Body)
	if body != "" {
		sb.WriteString("\n")
		sb.WriteString(body)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// ExtractTags finds tag-like tokens (e.g. #work) in a markdown body text.
// It ignores headers (e.g. "# Heading").
func ExtractTags(body string) []string {
	lines := strings.Split(body, "\n")
	var tags []string
	seen := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Markdown headers have space, e.g. "# Heading"
			isHeader := false
			for i := 1; i <= 6; i++ {
				prefix := strings.Repeat("#", i) + " "
				if strings.HasPrefix(trimmed, prefix) {
					isHeader = true
					break
				}
			}
			if isHeader {
				continue
			}
		}

		matches := tagRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tag := strings.ToLower(match[1])
				// Ignore hex colors (e.g. #fff, #aabbcc)
				isHexColor := (len(tag) == 3 || len(tag) == 6) && regexp.MustCompile(`^[0-9a-f]+$`).MatchString(tag)
				if isHexColor {
					continue
				}
				if !seen[tag] {
					seen[tag] = true
					tags = append(tags, tag)
				}
			}
		}
	}

	return tags
}

// MergeTags merges front matter tags and extracted body tags, removing duplicates.
func MergeTags(frontMatterTags []string, bodyTags []string) []string {
	seen := make(map[string]bool)
	var merged []string

	for _, t := range frontMatterTags {
		clean := strings.ToLower(strings.TrimSpace(t))
		if clean != "" && !seen[clean] {
			seen[clean] = true
			merged = append(merged, clean)
		}
	}

	for _, t := range bodyTags {
		clean := strings.ToLower(strings.TrimSpace(t))
		if clean != "" && !seen[clean] {
			seen[clean] = true
			merged = append(merged, clean)
		}
	}

	return merged
}

func splitLines(body string) []string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	return strings.Split(body, "\n")
}

func normalizeDisplayLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, "#*- ")
	line = strings.Join(strings.Fields(line), " ")
	return line
}
