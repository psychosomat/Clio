package notes

import "strings"

// Matches checks if a Note meets the specified filtering/search criteria.
func Matches(note Note, opts ListOptions) bool {
	// 1. Archive filter: if note is archived, only show if specifically requested
	if note.Archived && !opts.IncludeArchived {
		return false
	}

	// 2. Query search: case-insensitive match on derived title or body
	if opts.Query != "" {
		query := strings.ToLower(strings.TrimSpace(opts.Query))

		if strings.Contains(strings.ToLower(note.DisplayTitle()), query) {
			return true
		}

		if strings.Contains(strings.ToLower(note.Body), query) {
			return true
		}

		return false
	}

	return true
}

// FilterNotes filters a list of Notes using Matches.
func FilterNotes(notes []Note, opts ListOptions) []Note {
	var filtered []Note
	for _, note := range notes {
		if Matches(note, opts) {
			filtered = append(filtered, note)
		}
	}
	return filtered
}
