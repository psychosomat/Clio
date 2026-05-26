package notes

import "testing"

func TestSearchAndFilter(t *testing.T) {
	notes := []Note{
		{
			ID:   "1",
			Body: "Work Meeting\n\nDiscussed the new Go project.",
		},
		{
			ID:   "2",
			Body: "Shopping List\n\nBuy apples, milk, and coffee.",
		},
		{
			ID:   "3",
			Body: "App Idea\n\nA lightweight terminal note-taking app written in Go.",
		},
		{
			ID:       "4",
			Body:     "Old Ideas\n\nSome old thoughts.",
			Archived: true,
		},
	}

	// 1. Basic active listing (no archived notes)
	opts := ListOptions{}
	res := FilterNotes(notes, opts)
	if len(res) != 3 {
		t.Errorf("expected 3 active notes, got %d", len(res))
	}

	// 2. Query search by title (case-insensitive)
	opts = ListOptions{Query: "work"}
	res = FilterNotes(notes, opts)
	if len(res) != 1 || res[0].ID != "1" {
		t.Errorf("expected only note 1 for query 'work', got %d results", len(res))
	}

	// 3. Query search by body (case-insensitive)
	opts = ListOptions{Query: "coffee"}
	res = FilterNotes(notes, opts)
	if len(res) != 1 || res[0].ID != "2" {
		t.Errorf("expected only note 2 for query 'coffee', got %d results", len(res))
	}

	// 4. Search by body-derived title
	opts = ListOptions{Query: "idea"}
	res = FilterNotes(notes, opts)
	if len(res) != 1 || res[0].ID != "3" {
		t.Errorf("expected note 3 for query 'idea', got %d results", len(res))
	}

	// 5. Include archived
	opts = ListOptions{IncludeArchived: true}
	res = FilterNotes(notes, opts)
	if len(res) != 4 {
		t.Errorf("expected 4 notes including archived, got %d", len(res))
	}
}
