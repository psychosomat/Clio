package app

import (
	"strings"
	"testing"

	"clio/internal/markdownpreview"
)

func TestRenderMarkdownPreviewSupportsCodeAndTable(t *testing.T) {
	renderer := markdownpreview.NewRenderer()
	body := "# Title\n\n```go\nfmt.Println(\"ok\")\n```\n\n| A | B |\n| --- | --- |\n| 1 | 2 |"

	result, err := renderer.Render(body, markdownpreview.RenderOptions{
		Width:  60,
		Height: 20,
	})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	rendered := result.ANSI
	if strings.TrimSpace(rendered) == "" {
		t.Fatal("expected markdown preview to render content")
	}
	if !strings.Contains(rendered, "Title") {
		t.Fatalf("expected rendered preview to contain heading, got %q", rendered)
	}
	if !strings.Contains(rendered, "Println") {
		t.Fatalf("expected rendered preview to contain rendered code block, got %q", rendered)
	}
	if !strings.Contains(rendered, "│") {
		t.Fatalf("expected rendered preview to contain rendered table, got %q", rendered)
	}
}
