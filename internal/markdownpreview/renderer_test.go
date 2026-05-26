package markdownpreview

import (
	"strings"
	"testing"
)

func TestRendererScenarios(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	tests := []struct {
		name       string
		markdown   string
		contains   []string
		linkKinds  []LinkKind
		blockKinds []BlockKind
		anchors    []string
	}{
		{
			name:       "headings and hr",
			markdown:   "# Title\n\n## Subtitle\n\n---",
			contains:   []string{"# Title", "## Subtitle", "─"},
			blockKinds: []BlockKind{BlockKindHR},
			anchors:    []string{"title", "subtitle"},
		},
		{
			name:       "math and footnotes",
			markdown:   "Inline $a^2$ foot[^1].\n\n$$\na^2+b^2\n$$\n\n[^1]: Footnote body",
			contains:   []string{"a²", "$$\na^2+b^2\n$$", "Footnotes", "Footnote body"},
			blockKinds: []BlockKind{BlockKindMath, BlockKindFootnotes},
			anchors:    []string{"fn:[1]", "fnref:[1]"},
		},
		{
			name:      "links and wiki",
			markdown:  "[site](https://example.com) [[Note#Section|Jump]] ![[asset.png]]",
			contains:  []string{"site", "Jump", "[Embed: asset.png]"},
			linkKinds: []LinkKind{LinkKindExternal, LinkKindWiki, LinkKindEmbed},
		},
		{
			name:       "callout and mermaid",
			markdown:   "> [!warning] Careful\n> body\n\n```mermaid\ngraph TD;\nA-->B;\n```",
			contains:   []string{"WARNING", "Careful", "Mermaid diagram"},
			blockKinds: []BlockKind{BlockKindQuote, BlockKindMermaid},
		},
		{
			name:       "page break",
			markdown:   "<div style=\"page-break-after: always\"></div>\n\n<a href=\"https://example.com\">link</a>",
			contains:   []string{"Page Break", "link"},
			blockKinds: []BlockKind{BlockKindPageBreak},
		},
		{
			name:     "table and raw html",
			markdown: "| A | B |\n| --- | --- |\n| 1 | 2 |\n\nLine with <code>x</code> and <strong>bold</strong>.",
			contains: []string{"│", "A", "B", "x", "bold"},
		},
		{
			name:      "raw html anchor inline",
			markdown:  "Inline <a href=\"https://example.com\">link</a>.",
			contains:  []string{"Inline", "link"},
			linkKinds: []LinkKind{LinkKindExternal},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render(tt.markdown, RenderOptions{Width: 72, TerminalLinks: true})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if strings.TrimSpace(result.ANSI) == "" {
				t.Fatal("expected rendered output to be non-empty")
			}
			for _, want := range tt.contains {
				if !strings.Contains(result.PlainText, want) && !strings.Contains(result.ANSI, want) {
					t.Fatalf("expected output to contain %q\nANSI:\n%s\n\nPlain:\n%s", want, result.ANSI, result.PlainText)
				}
			}
			for _, kind := range tt.linkKinds {
				if !hasLinkKind(result.Links, kind) {
					t.Fatalf("expected link kind %q in %+v", kind, result.Links)
				}
			}
			for _, kind := range tt.blockKinds {
				if !hasBlockKind(result.Blocks, kind) {
					t.Fatalf("expected block kind %q in %+v", kind, result.Blocks)
				}
			}
			for _, anchor := range tt.anchors {
				if _, ok := result.AnchorLines[anchor]; !ok {
					t.Fatalf("expected anchor %q in %+v", anchor, result.AnchorLines)
				}
			}
		})
	}
}

func TestRendererTruncatesToHeight(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	result, err := renderer.Render("# A\n\nB\n\nC\n\nD\n\nE", RenderOptions{Width: 40, Height: 3})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !result.Truncated {
		t.Fatal("expected truncated result")
	}
	if !strings.Contains(result.PlainText, "... (truncated)") {
		t.Fatalf("expected truncation marker, got %q", result.PlainText)
	}
}

func TestRendererHighlightsCode(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	result, err := renderer.Render("```go\nfmt.Println(\"ok\")\n```", RenderOptions{Width: 80})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(result.ANSI, "\x1b[") {
		t.Fatalf("expected ANSI-highlighted code, got %q", result.ANSI)
	}
}

func TestRendererBeautifiesMath(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	result, err := renderer.Render("Inline $x_i^2 + \\alpha$ and\n\n$$\n\\sum_{i=1}^{n} x_i\n$$", RenderOptions{Width: 80})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(result.ANSI, "α") || !strings.Contains(result.ANSI, "²") {
		t.Fatalf("expected beautified inline math, got %q", result.ANSI)
	}
	if !strings.Contains(result.ANSI, "∑") {
		t.Fatalf("expected beautified block math, got %q", result.ANSI)
	}
}

func hasLinkKind(links []LinkTarget, kind LinkKind) bool {
	for _, link := range links {
		if link.Kind == kind {
			return true
		}
	}
	return false
}

func hasBlockKind(blocks []BlockMeta, kind BlockKind) bool {
	for _, block := range blocks {
		if block.Kind == kind {
			return true
		}
	}
	return false
}
