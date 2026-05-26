package markdownpreview

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	mathjax "github.com/litao91/goldmark-mathjax"
	obsidian "github.com/powerman/goldmark-obsidian"
	obsast "github.com/powerman/goldmark-obsidian/obsast"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	ast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/wikilink"
	"golang.org/x/net/html"
)

type LinkKind string

const (
	LinkKindExternal LinkKind = "external"
	LinkKindLocal    LinkKind = "local-file"
	LinkKindWiki     LinkKind = "wiki"
	LinkKindAnchor   LinkKind = "anchor"
	LinkKindEmbed    LinkKind = "embed"
)

type BlockKind string

const (
	BlockKindMath      BlockKind = "math"
	BlockKindMermaid   BlockKind = "mermaid"
	BlockKindHTML      BlockKind = "html"
	BlockKindEmbed     BlockKind = "embed"
	BlockKindFootnotes BlockKind = "footnotes"
	BlockKindCode      BlockKind = "code"
	BlockKindQuote     BlockKind = "quote"
	BlockKindHR        BlockKind = "hr"
	BlockKindPageBreak BlockKind = "pagebreak"
)

type ThemeConfig struct{}

type RenderOptions struct {
	Width           int
	Height          int
	TerminalLinks   bool
	Theme           ThemeConfig
	ActiveLinkIndex int
}

type LinkTarget struct {
	Index     int
	Label     string
	URL       string
	Kind      LinkKind
	StartLine int
	EndLine   int
}

type BlockMeta struct {
	Kind         BlockKind
	Source       string
	DisplayLabel string
	StartLine    int
	EndLine      int
}

type RenderResult struct {
	ANSI        string
	PlainText   string
	Links       []LinkTarget
	Blocks      []BlockMeta
	Truncated   bool
	AnchorLines map[string]int
}

type Renderer struct {
	md     goldmark.Markdown
	styles styles
}

type styles struct {
	paragraph    lipgloss.Style
	muted        lipgloss.Style
	heading      [6]lipgloss.Style
	inlineCode   lipgloss.Style
	codeBlock    lipgloss.Style
	link         lipgloss.Style
	activeLink   lipgloss.Style
	math         lipgloss.Style
	mathBlock    lipgloss.Style
	blockquote   lipgloss.Style
	callout      lipgloss.Style
	calloutTitle lipgloss.Style
	hr           lipgloss.Style
	pageBreak    lipgloss.Style
	footnote     lipgloss.Style
	embed        lipgloss.Style
	html         lipgloss.Style
	tableHeader  lipgloss.Style
	tableBorder  lipgloss.Style
	listBullet   lipgloss.Style
	tag          lipgloss.Style
}

var (
	inlineFootnoteRE     = regexp.MustCompile(`\^\[([^\]]+)\]`)
	pageBreakRE          = regexp.MustCompile(`(?i)(page-break-(after|before)\s*:\s*always|class\s*=\s*"page-break"|class\s*=\s*'page-break')`)
	calloutRE            = regexp.MustCompile(`^\[!([a-z0-9_-]+)\]([+-])?\s*(.*)$`)
	htmlAnchorOpenRE     = regexp.MustCompile(`(?i)<a\s+[^>]*href\s*=\s*['"]([^'"]+)['"][^>]*>`)
	blockMathFallbackRE  = regexp.MustCompile(`(?ms)(^|\n)\$\$\s*\n?(.*?)\n?\$\$(\n|$)`)
	inlineMathFallbackRE = regexp.MustCompile(`\$([^\$\n]+)\$`)
	superscriptRE        = regexp.MustCompile(`\^\{([^}]+)\}|\^([A-Za-z0-9+\-=()])`)
	subscriptRE          = regexp.MustCompile(`_\{([^}]+)\}|_([A-Za-z0-9+\-=()])`)
)

var mathUnicodeMap = map[string]string{
	`\alpha`: "α", `\beta`: "β", `\gamma`: "γ", `\delta`: "δ", `\epsilon`: "ε",
	`\theta`: "θ", `\lambda`: "λ", `\mu`: "μ", `\pi`: "π", `\sigma`: "σ",
	`\phi`: "φ", `\omega`: "ω", `\Gamma`: "Γ", `\Delta`: "Δ", `\Theta`: "Θ",
	`\Lambda`: "Λ", `\Pi`: "Π", `\Sigma`: "Σ", `\Phi`: "Φ", `\Omega`: "Ω",
	`\sum`: "∑", `\prod`: "∏", `\int`: "∫",
	`\partial`: "∂", `\infty`: "∞", `\nabla`: "∇", `\forall`: "∀", `\exists`: "∃",
	`\in`: "∈", `\notin`: "∉", `\subset`: "⊂", `\supset`: "⊃", `\subseteq`: "⊆",
	`\supseteq`: "⊇", `\cup`: "∪", `\cap`: "∩", `\times`: "×", `\cdot`: "·",
	`\pm`: "±", `\neq`: "≠", `\leq`: "≤", `\geq`: "≥", `\approx`: "≈",
	`\rightarrow`: "→", `\leftarrow`: "←", `\leftrightarrow`: "↔", `\to`: "→",
}

var superscriptMap = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴', '5': '⁵', '6': '⁶', '7': '⁷',
	'8': '⁸', '9': '⁹', '+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
	'n': 'ⁿ', 'i': 'ⁱ', 'x': 'ˣ', 'y': 'ʸ', 'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ',
	'e': 'ᵉ', 'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'j': 'ʲ', 'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ',
	'o': 'ᵒ', 'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ', 'v': 'ᵛ', 'w': 'ʷ',
	'z': 'ᶻ',
}

var subscriptMap = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄', '5': '₅', '6': '₆', '7': '₇',
	'8': '₈', '9': '₉', '+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ', 'k': 'ₖ', 'l': 'ₗ',
	'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ', 'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ',
	'u': 'ᵤ', 'v': 'ᵥ', 'x': 'ₓ',
}

func NewRenderer() *Renderer {
	return &Renderer{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Footnote,
				extension.Linkify,
				emoji.Emoji,
				obsidian.NewPlugTasks(),
				obsidian.NewObsidian(),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
		styles: styles{
			paragraph:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")),
			muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")),
			inlineCode: lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")).Background(lipgloss.Color("#1f2330")).Padding(0, 1),
			codeBlock:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")).Background(lipgloss.Color("#1f2330")).Padding(0, 1),
			link:       lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Underline(true),
			activeLink: lipgloss.NewStyle().Foreground(lipgloss.Color("#1e1e2e")).Background(lipgloss.Color("#8be9fd")).Bold(true).Underline(true),
			math:       lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Italic(true),
			mathBlock:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Background(lipgloss.Color("#1f2330")).Padding(0, 1),
			blockquote: lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")).BorderForeground(lipgloss.Color("#6272a4")),
			callout:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")).Background(lipgloss.Color("#303448")).Padding(0, 1),
			calloutTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1e1e2e")).
				Background(lipgloss.Color("#8be9fd")).
				Bold(true).
				Padding(0, 1),
			hr:          lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")),
			pageBreak:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Bold(true),
			footnote:    lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")),
			embed:       lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Background(lipgloss.Color("#1f2330")).Padding(0, 1),
			html:        lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")),
			tableHeader: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8be9fd")),
			tableBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")),
			listBullet:  lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true),
			tag:         lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Italic(true),
		},
	}
}

func (r *Renderer) Render(markdown string, opts RenderOptions) (RenderResult, error) {
	if opts.Width < 20 {
		opts.Width = 20
	}
	source := preprocessMathFallbacks(preprocessInlineFootnotes(strings.ReplaceAll(markdown, "\f", "\n<div class=\"page-break\"></div>\n")))
	doc := r.md.Parser().Parse(text.NewReader([]byte(source)))
	ctx := &renderContext{
		renderer: r,
		source:   []byte(source),
		opts:     opts,
		result: RenderResult{
			AnchorLines: make(map[string]int),
		},
	}
	r.renderChildren(doc, ctx, 0)
	ctx.flushParagraph()
	ansi := strings.TrimRight(ctx.out.String(), "\n")
	plain := strings.TrimRight(ctx.plain.String(), "\n")
	if opts.Height > 0 {
		ansi, plain, ctx.result.Truncated = truncateToHeight(ansi, plain, opts.Height, r.styles.muted.Render("... (truncated)"))
	}
	ctx.result.ANSI = ansi
	ctx.result.PlainText = plain
	return ctx.result, nil
}

type renderContext struct {
	renderer *Renderer
	source   []byte
	opts     RenderOptions
	result   RenderResult
	out      strings.Builder
	plain    strings.Builder
}

func (c *renderContext) line() int {
	if c.out.Len() == 0 {
		return 0
	}
	return strings.Count(c.out.String(), "\n")
}

func (c *renderContext) writeBlock(text, plain string) {
	if text == "" {
		return
	}
	if c.out.Len() > 0 && !strings.HasSuffix(c.out.String(), "\n\n") {
		if strings.HasSuffix(c.out.String(), "\n") {
			c.out.WriteString("\n")
			c.plain.WriteString("\n")
		} else {
			c.out.WriteString("\n\n")
			c.plain.WriteString("\n\n")
		}
	}
	c.out.WriteString(text)
	c.plain.WriteString(plain)
	if !strings.HasSuffix(text, "\n") {
		c.out.WriteString("\n")
	}
	if !strings.HasSuffix(plain, "\n") {
		c.plain.WriteString("\n")
	}
}

func (c *renderContext) addBlockMeta(kind BlockKind, source, label string, startLine int) {
	endLine := c.line()
	c.result.Blocks = append(c.result.Blocks, BlockMeta{
		Kind:         kind,
		Source:       strings.TrimSpace(source),
		DisplayLabel: label,
		StartLine:    startLine,
		EndLine:      endLine,
	})
}

func (c *renderContext) flushParagraph() {}

func (r *Renderer) renderChildren(parent ast.Node, c *renderContext, depth int) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderBlock(child, c, depth)
	}
}

func (r *Renderer) renderBlock(node ast.Node, c *renderContext, depth int) {
	switch n := node.(type) {
	case *ast.Heading:
		start := c.line()
		title := r.renderInlineChildren(n, c)
		level := n.Level
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		prefix := strings.Repeat("#", level) + " "
		rendered := r.styles.headingStyle(level).Render(prefix + title)
		c.writeBlock(rendered, prefix+stripANSI(title))
		anchor := headingAnchor(n, c.source, title)
		c.result.AnchorLines[anchor] = start
	case *ast.Paragraph:
		start := c.line()
		rendered, plain := r.renderParagraph(n, c, depth)
		c.writeBlock(rendered, plain)
		if calloutKind, ok := detectCallout(n, c); ok {
			c.addBlockMeta(BlockKindQuote, plain, calloutKind, start)
		}
	case *ast.Blockquote:
		start := c.line()
		rendered, plain, label := r.renderBlockquote(n, c, depth)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindQuote, plain, label, start)
	case *ast.List:
		rendered, plain := r.renderList(n, c, depth)
		c.writeBlock(rendered, plain)
	case *extast.Table:
		rendered, plain := r.renderTable(n, c)
		c.writeBlock(rendered, plain)
	case *ast.FencedCodeBlock:
		start := c.line()
		rendered, plain := r.renderFencedCode(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindCode, plain, "Code", start)
	case *ast.CodeBlock:
		start := c.line()
		rendered, plain := r.renderCodeBlock(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindCode, plain, "Code", start)
	case *ast.HTMLBlock:
		start := c.line()
		rendered, plain, kind, label := r.renderHTMLBlock(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(kind, plain, label, start)
	case *ast.ThematicBreak:
		start := c.line()
		rule := r.styles.hr.Render(strings.Repeat("─", c.opts.Width))
		c.writeBlock(rule, strings.Repeat("-", c.opts.Width))
		c.addBlockMeta(BlockKindHR, strings.Repeat("-", c.opts.Width), "Horizontal Rule", start)
	case *mathjax.MathBlock:
		start := c.line()
		rendered, plain := r.renderMathBlock(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindMath, plain, "Math", start)
	case *extast.FootnoteList:
		start := c.line()
		c.result.AnchorLines["footnotes"] = start
		rendered, plain := r.renderFootnoteList(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindFootnotes, plain, "Footnotes", start)
	case *mermaid.Block:
		start := c.line()
		rendered, plain := r.renderMermaidBlock(n, c)
		c.writeBlock(rendered, plain)
		c.addBlockMeta(BlockKindMermaid, plain, "Mermaid diagram", start)
	case *ast.TextBlock:
		rendered := r.renderInlineChildren(n, c)
		c.writeBlock(rendered, stripANSI(rendered))
	default:
		if node.Type() == ast.TypeBlock {
			r.renderChildren(node, c, depth+1)
		}
	}
}

func (r *Renderer) renderParagraph(n *ast.Paragraph, c *renderContext, depth int) (string, string) {
	text := r.renderInlineChildren(n, c)
	plain := stripANSI(text)
	wrapped := lipgloss.NewStyle().Width(c.opts.Width - depth*2).Render(text)
	return indentBlock(wrapped, depth*2), indentBlock(plain, depth*2)
}

func (r *Renderer) renderBlockquote(n *ast.Blockquote, c *renderContext, depth int) (string, string, string) {
	var innerANSI []string
	var innerPlain []string
	var label string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch p := child.(type) {
		case *ast.Paragraph:
			text := r.renderInlineChildren(p, c)
			plain := stripANSI(text)
			if kind, ok := detectCallout(p, c); ok {
				label = kind
				title, body := splitCallout(plain)
				header := r.styles.calloutTitle.Render(strings.ToUpper(kind))
				content := r.styles.callout.Render(title)
				if body != "" {
					content += "\n" + r.styles.callout.Render(body)
				}
				return header + "\n" + content, strings.ToUpper(kind) + "\n" + title + "\n" + body, kind
			}
			innerANSI = append(innerANSI, text)
			innerPlain = append(innerPlain, plain)
		default:
			sub := &renderContext{renderer: c.renderer, source: c.source, opts: c.opts, result: RenderResult{AnchorLines: c.result.AnchorLines}}
			r.renderBlock(child, sub, depth+1)
			innerANSI = append(innerANSI, strings.TrimSpace(sub.out.String()))
			innerPlain = append(innerPlain, strings.TrimSpace(sub.plain.String()))
		}
	}
	bodyANSI := strings.Join(innerANSI, "\n")
	bodyPlain := strings.Join(innerPlain, "\n")
	lines := strings.Split(bodyANSI, "\n")
	for i, line := range lines {
		lines[i] = r.styles.muted.Render("│ ") + line
	}
	plainLines := strings.Split(bodyPlain, "\n")
	for i, line := range plainLines {
		plainLines[i] = "│ " + line
	}
	return strings.Join(lines, "\n"), strings.Join(plainLines, "\n"), label
}

func (r *Renderer) renderList(n *ast.List, c *renderContext, depth int) (string, string) {
	var ansi []string
	var plain []string
	index := n.Start
	if index == 0 {
		index = 1
	}
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		bullet := "•"
		if n.IsOrdered() {
			bullet = fmt.Sprintf("%d.", index)
			index++
		}
		itemANSI, itemPlain := r.renderListItem(item, c, depth+1)
		lines := strings.Split(itemANSI, "\n")
		plainLines := strings.Split(itemPlain, "\n")
		prefix := r.styles.listBullet.Render(bullet + " ")
		for i := range lines {
			if i == 0 {
				lines[i] = strings.Repeat(" ", depth*2) + prefix + lines[i]
				plainLines[i] = strings.Repeat(" ", depth*2) + bullet + " " + plainLines[i]
				continue
			}
			lines[i] = strings.Repeat(" ", depth*2+2) + lines[i]
			plainLines[i] = strings.Repeat(" ", depth*2+2) + plainLines[i]
		}
		ansi = append(ansi, strings.Join(lines, "\n"))
		plain = append(plain, strings.Join(plainLines, "\n"))
	}
	return strings.Join(ansi, "\n"), strings.Join(plain, "\n")
}

func (r *Renderer) renderListItem(item ast.Node, c *renderContext, depth int) (string, string) {
	var ansi []string
	var plain []string
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Paragraph:
			a, p := r.renderParagraph(n, c, depth)
			ansi = append(ansi, strings.TrimSpace(a))
			plain = append(plain, strings.TrimSpace(p))
		case *ast.List:
			a, p := r.renderList(n, c, depth+1)
			ansi = append(ansi, a)
			plain = append(plain, p)
		default:
			sub := &renderContext{renderer: c.renderer, source: c.source, opts: c.opts, result: RenderResult{AnchorLines: c.result.AnchorLines}}
			r.renderBlock(child, sub, depth)
			ansi = append(ansi, strings.TrimSpace(sub.out.String()))
			plain = append(plain, strings.TrimSpace(sub.plain.String()))
		}
	}
	return strings.Join(ansi, "\n"), strings.Join(plain, "\n")
}

func (r *Renderer) renderTable(n *extast.Table, c *renderContext) (string, string) {
	var rows [][]string
	var plainRows [][]string
	var aligns []extast.Alignment
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch tr := row.(type) {
		case *extast.TableHeader:
			var headerVals []string
			var headerPlain []string
			for cell := tr.FirstChild(); cell != nil; cell = cell.NextSibling() {
				val := r.renderInlineChildren(cell, c)
				headerVals = append(headerVals, val)
				headerPlain = append(headerPlain, stripANSI(val))
			}
			rows = append(rows, headerVals)
			plainRows = append(plainRows, headerPlain)
		case *extast.TableRow:
			aligns = tr.Alignments
			var rowVals []string
			var rowPlain []string
			for cell := tr.FirstChild(); cell != nil; cell = cell.NextSibling() {
				val := r.renderInlineChildren(cell, c)
				rowVals = append(rowVals, val)
				rowPlain = append(rowPlain, stripANSI(val))
			}
			rows = append(rows, rowVals)
			plainRows = append(plainRows, rowPlain)
		}
	}
	if len(rows) == 0 {
		return "", ""
	}
	widths := tableWidths(plainRows, c.opts.Width)
	var ansiLines []string
	var plainLines []string
	border := tableBorder(widths)
	ansiLines = append(ansiLines, r.styles.tableBorder.Render(border))
	plainLines = append(plainLines, border)
	for i, row := range rows {
		line, plain := formatTableRow(row, plainRows[i], widths, aligns, i == 0, r.styles.tableHeader)
		ansiLines = append(ansiLines, r.styles.tableBorder.Render("│")+" "+line+" "+r.styles.tableBorder.Render("│"))
		plainLines = append(plainLines, "│ "+plain+" │")
		if i == 0 {
			ansiLines = append(ansiLines, r.styles.tableBorder.Render(border))
			plainLines = append(plainLines, border)
		}
	}
	ansiLines = append(ansiLines, r.styles.tableBorder.Render(border))
	plainLines = append(plainLines, border)
	return strings.Join(ansiLines, "\n"), strings.Join(plainLines, "\n")
}

func (r *Renderer) renderFencedCode(n *ast.FencedCodeBlock, c *renderContext) (string, string) {
	var lines []string
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		lines = append(lines, strings.TrimRight(string(seg.Value(c.source)), "\n"))
	}
	info := strings.TrimSpace(string(n.Language(c.source)))
	header := "```"
	if info != "" {
		header += info
	}
	plain := header + "\n" + strings.Join(lines, "\n") + "\n```"
	highlighted := highlightCode(strings.Join(lines, "\n"), info)
	rendered := header + "\n" + highlighted + "\n```"
	return r.styles.codeBlock.Render(rendered), plain
}

func (r *Renderer) renderCodeBlock(n *ast.CodeBlock, c *renderContext) (string, string) {
	var lines []string
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		lines = append(lines, strings.TrimRight(string(seg.Value(c.source)), "\n"))
	}
	plain := "```\n" + strings.Join(lines, "\n") + "\n```"
	return r.styles.codeBlock.Render(plain), plain
}

func (r *Renderer) renderMathBlock(n *mathjax.MathBlock, c *renderContext) (string, string) {
	var body strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		body.Write(seg.Value(c.source))
	}
	raw := strings.TrimRight(body.String(), "\n")
	pretty := beautifyMath(raw)
	plain := "$$\n" + raw + "\n$$"
	rendered := pretty
	if pretty != raw {
		rendered += "\n" + r.styles.muted.Render("TeX: "+raw)
	}
	return r.styles.mathBlock.Render(rendered), plain
}

func (r *Renderer) renderFootnoteList(n *extast.FootnoteList, c *renderContext) (string, string) {
	var ansi []string
	var plain []string
	ansi = append(ansi, r.styles.footnote.Render("Footnotes"))
	plain = append(plain, "Footnotes")
	for fn := n.FirstChild(); fn != nil; fn = fn.NextSibling() {
		f, ok := fn.(*extast.Footnote)
		if !ok {
			continue
		}
		idx := fmt.Sprintf("[%d]", f.Index)
		c.result.AnchorLines["fn:"+idx] = c.line() + len(ansi)
		text := strings.TrimSpace(r.renderInlineChildren(f, c))
		backref := r.renderRegisteredLink("↩", "#fnref:"+idx, LinkKindAnchor, c)
		ansi = append(ansi, r.styles.footnote.Render(idx)+" "+text+" "+backref)
		plain = append(plain, idx+" "+stripANSI(text)+" ↩")
	}
	return strings.Join(ansi, "\n"), strings.Join(plain, "\n")
}

func (r *Renderer) renderMermaidBlock(n *mermaid.Block, c *renderContext) (string, string) {
	var src strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		src.Write(seg.Value(c.source))
	}
	body := strings.TrimSpace(src.String())
	snippet := body
	if len(snippet) > 180 {
		snippet = snippet[:180] + "..."
	}
	plain := "[Mermaid diagram]\n" + snippet
	return r.styles.embed.Render(plain), plain
}

func (r *Renderer) renderHTMLBlock(n *ast.HTMLBlock, c *renderContext) (string, string, BlockKind, string) {
	var src strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		src.Write(seg.Value(c.source))
	}
	raw := strings.TrimSpace(src.String())
	if strings.HasPrefix(strings.ToLower(raw), "<clio-math-block>") && strings.HasSuffix(strings.ToLower(raw), "</clio-math-block>") {
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "<clio-math-block>"), "</clio-math-block>"))
		plain := "$$\n" + body + "\n$$"
		rendered := beautifyMath(body)
		if rendered != body {
			rendered += "\n" + r.styles.muted.Render("TeX: "+body)
		}
		return r.styles.mathBlock.Render(rendered), plain, BlockKindMath, "Math"
	}
	if pageBreakRE.MatchString(raw) {
		rest := strings.TrimSpace(stripPageBreakMarkup(raw))
		rendered := r.styles.pageBreak.Render("──────── Page Break ────────")
		plain := "Page Break"
		if rest != "" {
			htmlRendered, htmlPlain := r.renderHTMLFragment(rest, c)
			if htmlPlain != "" {
				rendered += "\n" + htmlRendered
				plain += "\n" + htmlPlain
			}
		}
		return rendered, plain, BlockKindPageBreak, "Page Break"
	}
	if strings.Contains(strings.ToLower(raw), "<hr") {
		rendered := r.styles.hr.Render(strings.Repeat("─", c.opts.Width))
		return rendered, strings.Repeat("-", c.opts.Width), BlockKindHR, "Horizontal Rule"
	}
	rendered, plain := r.renderHTMLFragment(raw, c)
	if strings.TrimSpace(plain) == "" {
		plain = "[HTML block]\n" + raw
		rendered = r.styles.html.Render(plain)
	}
	return rendered, plain, BlockKindHTML, "HTML"
}

func (r *Renderer) renderInlineChildren(parent ast.Node, c *renderContext) string {
	var out strings.Builder
	activeHTMLLink := ""
	activeInlineMath := false
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			text := string(n.Segment.Value(c.source))
			switch {
			case activeInlineMath:
				out.WriteString(r.styles.math.Render(beautifyMath(text)))
			case activeHTMLLink != "":
				out.WriteString(r.renderRegisteredLink(text, activeHTMLLink, inferLinkKind(activeHTMLLink), c))
			default:
				out.WriteString(text)
			}
			if n.HardLineBreak() || n.SoftLineBreak() {
				out.WriteByte('\n')
			}
		case *ast.String:
			out.Write(n.Value)
		case *ast.CodeSpan:
			out.WriteString(r.styles.inlineCode.Render(extractInlineText(n, c.source)))
		case *ast.Emphasis:
			content := r.renderInlineChildren(n, c)
			if n.Level == 2 {
				out.WriteString(lipgloss.NewStyle().Bold(true).Render(content))
			} else {
				out.WriteString(lipgloss.NewStyle().Italic(true).Render(content))
			}
		case *extast.Strikethrough:
			out.WriteString(lipgloss.NewStyle().Strikethrough(true).Render(r.renderInlineChildren(n, c)))
		case *ast.Link:
			label := r.renderInlineChildren(n, c)
			out.WriteString(r.renderRegisteredLink(label, string(n.Destination), inferLinkKind(string(n.Destination)), c))
		case *ast.AutoLink:
			url := string(n.URL(c.source))
			out.WriteString(r.renderRegisteredLink(url, url, LinkKindExternal, c))
		case *wikilink.Node:
			label := strings.TrimSpace(r.renderInlineChildren(n, c))
			target := string(n.Target)
			if label == "" {
				label = target
				if frag := strings.TrimSpace(string(n.Fragment)); frag != "" {
					label = target + "#" + frag
				}
			}
			kind := LinkKindWiki
			if n.Embed {
				kind = LinkKindEmbed
			}
			url := target
			if frag := strings.TrimSpace(string(n.Fragment)); frag != "" {
				url += "#" + frag
			}
			if kind == LinkKindEmbed {
				out.WriteString(r.styles.embed.Render("[Embed: " + label + "]"))
				r.registerLink(label, url, kind, c)
				continue
			}
			out.WriteString(r.renderRegisteredLink(label, url, kind, c))
		case *ast.Image:
			label := extractInlineText(n, c.source)
			if label == "" {
				label = "image"
			}
			placeholder := "[Image: " + label + "]"
			out.WriteString(r.styles.embed.Render(placeholder))
			r.registerLink(label, string(n.Destination), LinkKindEmbed, c)
		case *extast.FootnoteLink:
			label := fmt.Sprintf("[%d]", n.Index)
			target := "#fn:" + label
			c.result.AnchorLines["fnref:"+label] = c.line()
			out.WriteString(r.renderRegisteredLink(label, target, LinkKindAnchor, c))
		case *extast.FootnoteBacklink:
			out.WriteString(r.renderRegisteredLink("↩", "#fnref:[1]", LinkKindAnchor, c))
		case *extast.TaskCheckBox:
			if n.IsChecked {
				out.WriteString("[x] ")
			} else {
				out.WriteString("[ ] ")
			}
		case *obsast.BlockID:
			anchor := "^" + string(n.ID)
			c.result.AnchorLines[anchor] = c.line()
			out.WriteString(r.styles.muted.Render(" " + anchor))
		case *obsast.PlugTasksStatus:
			out.WriteString("[" + string(n.Symbol) + "] ")
		case *mathjax.InlineMath:
			content := strings.TrimSpace(extractInlineText(n, c.source))
			out.WriteString(r.styles.math.Render(beautifyMath(content)))
		case *ast.RawHTML:
			raw := extractInlineText(n, c.source)
			switch strings.ToLower(strings.TrimSpace(raw)) {
			case "<clio-math-inline>":
				activeInlineMath = true
				continue
			case "</clio-math-inline>":
				activeInlineMath = false
				continue
			}
			if href, open := htmlAnchorHref(raw); open {
				activeHTMLLink = href
				continue
			}
			if isHTMLAnchorClose(raw) {
				activeHTMLLink = ""
				continue
			}
			rendered, _ := r.renderHTMLFragment(raw, c)
			out.WriteString(rendered)
		default:
			if child.HasChildren() {
				out.WriteString(r.renderInlineChildren(child, c))
			}
		}
	}
	return out.String()
}

func (r *Renderer) registerLink(label, url string, kind LinkKind, c *renderContext) {
	index := len(c.result.Links)
	c.result.Links = append(c.result.Links, LinkTarget{
		Index:     index,
		Label:     label,
		URL:       url,
		Kind:      kind,
		StartLine: c.line(),
		EndLine:   c.line(),
	})
}

func (r *Renderer) renderRegisteredLink(label, url string, kind LinkKind, c *renderContext) string {
	index := len(c.result.Links)
	c.result.Links = append(c.result.Links, LinkTarget{
		Index:     index,
		Label:     stripANSI(label),
		URL:       url,
		Kind:      kind,
		StartLine: c.line(),
		EndLine:   c.line(),
	})
	style := r.styles.link
	if index == c.opts.ActiveLinkIndex {
		style = r.styles.activeLink
	}
	if c.opts.TerminalLinks && kind != LinkKindWiki && kind != LinkKindAnchor && kind != LinkKindEmbed {
		return style.Render(terminalHyperlink(url, stripANSI(label)))
	}
	return style.Render(label)
}

func (r *Renderer) renderHTMLFragment(raw string, c *renderContext) (string, string) {
	nodes, err := html.ParseFragment(strings.NewReader(raw), &html.Node{Type: html.ElementNode, Data: "div"})
	if err != nil {
		return r.styles.html.Render("[HTML] " + raw), "[HTML] " + raw
	}
	var ansi []string
	var plain []string
	for _, node := range nodes {
		a, p := r.renderHTMLNode(node, c)
		ansi = append(ansi, a)
		plain = append(plain, p)
	}
	return strings.TrimSpace(strings.Join(ansi, "")), strings.TrimSpace(strings.Join(plain, ""))
}

func (r *Renderer) renderHTMLNode(n *html.Node, c *renderContext) (string, string) {
	switch n.Type {
	case html.TextNode:
		return n.Data, n.Data
	case html.ElementNode:
		switch strings.ToLower(n.Data) {
		case "clio-math-inline":
			text := strings.TrimSpace(collectHTMLText(n))
			return r.styles.math.Render(beautifyMath(text)), text
		case "clio-math-block":
			text := strings.TrimSpace(collectHTMLText(n))
			rendered := beautifyMath(text)
			if rendered != text {
				rendered += "\n" + r.styles.muted.Render("TeX: "+text)
			}
			return r.styles.mathBlock.Render(rendered), text
		case "br":
			return "\n", "\n"
		case "hr":
			line := strings.Repeat("─", c.opts.Width)
			return r.styles.hr.Render(line), line
		case "a":
			href := attr(n, "href")
			label := collectHTMLText(n)
			if label == "" {
				label = href
			}
			return r.renderRegisteredLink(label, href, inferLinkKind(href), c), label
		case "strong", "b":
			text := collectHTMLText(n)
			return lipgloss.NewStyle().Bold(true).Render(text), text
		case "em", "i":
			text := collectHTMLText(n)
			return lipgloss.NewStyle().Italic(true).Render(text), text
		case "del":
			text := collectHTMLText(n)
			return lipgloss.NewStyle().Strikethrough(true).Render(text), text
		case "code", "kbd":
			text := collectHTMLText(n)
			return r.styles.inlineCode.Render(text), text
		case "sup", "sub", "mark":
			text := collectHTMLText(n)
			return r.styles.html.Render(text), text
		case "img":
			src := attr(n, "src")
			alt := attr(n, "alt")
			if alt == "" {
				alt = "image"
			}
			label := "[Image: " + alt + "]"
			r.registerLink(alt, src, LinkKindEmbed, c)
			return r.styles.embed.Render(label), label
		case "details":
			text := collectHTMLText(n)
			label := "[Details] " + strings.TrimSpace(text)
			return r.styles.html.Render(label), label
		case "summary", "div", "span", "p", "table", "thead", "tbody", "tr", "th", "td":
			text := strings.TrimSpace(collectHTMLText(n))
			return text, text
		default:
			text := strings.TrimSpace(collectHTMLText(n))
			if text == "" {
				text = "<" + n.Data + ">"
			}
			label := "[HTML " + strings.ToUpper(n.Data) + "] " + text
			return r.styles.html.Render(label), label
		}
	}
	return "", ""
}

func (s styles) headingStyle(level int) lipgloss.Style {
	switch level {
	case 1:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")).Bold(true)
	case 2:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true)
	case 3:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
	case 4:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")).Bold(true)
	case 5:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true)
	}
}

func preprocessInlineFootnotes(markdown string) string {
	counter := 0
	return inlineFootnoteRE.ReplaceAllStringFunc(markdown, func(match string) string {
		sub := inlineFootnoteRE.FindStringSubmatch(match)
		if len(sub) != 2 {
			return match
		}
		counter++
		id := fmt.Sprintf("inline-%d", counter)
		return fmt.Sprintf("[^%s]\n\n[^%s]: %s", id, id, sub[1])
	})
}

func preprocessMathFallbacks(markdown string) string {
	if strings.Contains(markdown, "$$") {
		markdown = blockMathFallbackRE.ReplaceAllStringFunc(markdown, func(match string) string {
			if strings.Contains(match, "<clio-math-block>") {
				return match
			}
			sub := blockMathFallbackRE.FindStringSubmatch(match)
			if len(sub) != 4 {
				return match
			}
			body := strings.TrimSpace(sub[2])
			return sub[1] + "<clio-math-block>\n" + body + "\n</clio-math-block>" + sub[3]
		})
	}
	return inlineMathFallbackRE.ReplaceAllStringFunc(markdown, func(match string) string {
		if strings.Contains(match, `\$`) {
			return match
		}
		sub := inlineMathFallbackRE.FindStringSubmatch(match)
		if len(sub) != 2 {
			return match
		}
		return "<clio-math-inline>" + sub[1] + "</clio-math-inline>"
	})
}

func truncateToHeight(ansi, plain string, height int, marker string) (string, string, bool) {
	ansiLines := strings.Split(ansi, "\n")
	plainLines := strings.Split(plain, "\n")
	if len(ansiLines) <= height {
		return ansi, plain, false
	}
	if height <= 1 {
		return marker, "... (truncated)", true
	}
	return strings.Join(append(ansiLines[:height-1], marker), "\n"), strings.Join(append(plainLines[:height-1], "... (truncated)"), "\n"), true
}

func highlightCode(source, language string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return source
	}
	if lexers.Get(language) == nil {
		return source
	}
	var buf bytes.Buffer
	if err := quick.Highlight(&buf, source, language, "terminal16m", "dracula"); err != nil {
		return source
	}
	return strings.TrimRight(buf.String(), "\n")
}

func beautifyMath(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return expr
	}
	for cmd, uni := range mathUnicodeMap {
		expr = strings.ReplaceAll(expr, cmd, uni)
	}
	expr = superscriptRE.ReplaceAllStringFunc(expr, func(m string) string {
		sub := superscriptRE.FindStringSubmatch(m)
		val := firstNonEmpty(sub[1], sub[2])
		return mapRunes(val, superscriptMap)
	})
	expr = subscriptRE.ReplaceAllStringFunc(expr, func(m string) string {
		sub := subscriptRE.FindStringSubmatch(m)
		val := firstNonEmpty(sub[1], sub[2])
		return mapRunes(val, subscriptMap)
	})
	expr = strings.NewReplacer(`\(`, "(", `\)`, ")", `\[`, "[", `\]`, "]").Replace(expr)
	return expr
}

func tableWidths(rows [][]string, width int) []int {
	cols := 0
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return nil
	}
	widths := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			cellWidth := lipgloss.Width(cell)
			if cellWidth > widths[i] {
				widths[i] = cellWidth
			}
		}
	}
	totalPadding := cols*3 + 1
	maxWidth := width - totalPadding
	if maxWidth < cols*8 {
		maxWidth = cols * 8
	}
	total := 0
	for _, w := range widths {
		if w < 8 {
			w = 8
		}
		total += w
	}
	if total <= maxWidth {
		for i, w := range widths {
			if w < 8 {
				widths[i] = 8
			}
		}
		return widths
	}
	shrink := total - maxWidth
	for shrink > 0 {
		changed := false
		for i := range widths {
			if widths[i] > 8 && shrink > 0 {
				widths[i]--
				shrink--
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return widths
}

func tableBorder(widths []int) string {
	var b strings.Builder
	b.WriteString("┼")
	for _, w := range widths {
		b.WriteString(strings.Repeat("─", w+2))
		b.WriteString("┼")
	}
	return b.String()
}

func formatTableRow(ansiCells, plainCells []string, widths []int, aligns []extast.Alignment, header bool, style lipgloss.Style) (string, string) {
	var ansi []string
	var plain []string
	for i, width := range widths {
		var a, p string
		if i < len(ansiCells) {
			a = ansiCells[i]
			p = plainCells[i]
		}
		paddedA := padCell(a, width, alignAt(aligns, i))
		paddedP := padCell(p, width, alignAt(aligns, i))
		if header {
			paddedA = style.Render(paddedA)
		}
		ansi = append(ansi, paddedA)
		plain = append(plain, paddedP)
	}
	return strings.Join(ansi, " │ "), strings.Join(plain, " │ ")
}

func alignAt(aligns []extast.Alignment, idx int) extast.Alignment {
	if idx < len(aligns) {
		return aligns[idx]
	}
	return extast.AlignLeft
}

func padCell(text string, width int, align extast.Alignment) string {
	plainWidth := lipgloss.Width(stripANSI(text))
	if plainWidth > width {
		text = lipgloss.NewStyle().MaxWidth(width).Render(text)
		plainWidth = lipgloss.Width(stripANSI(text))
	}
	padding := width - plainWidth
	switch align {
	case extast.AlignRight:
		return strings.Repeat(" ", padding) + text
	case extast.AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

func extractInlineText(n ast.Node, source []byte) string {
	if textNode, ok := n.(interface{ Text([]byte) []byte }); ok && !n.HasChildren() {
		return string(textNode.Text(source))
	}
	var out strings.Builder
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch t := child.(type) {
		case *ast.Text:
			out.Write(t.Segment.Value(source))
			if t.SoftLineBreak() || t.HardLineBreak() {
				out.WriteByte('\n')
			}
		case *ast.String:
			out.Write(t.Value)
		default:
			if child.HasChildren() {
				out.WriteString(extractInlineText(child, source))
			}
		}
	}
	return out.String()
}

func indentBlock(s string, spaces int) string {
	if spaces == 0 || s == "" {
		return s
	}
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func stripANSI(s string) string {
	var out bytes.Buffer
	inEsc := false
	oscMode := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == 0x1b {
			inEsc = true
			oscMode = false
			continue
		}
		if inEsc {
			if ch == ']' {
				oscMode = true
				continue
			}
			if oscMode {
				if ch == 0x07 {
					inEsc = false
					oscMode = false
				}
				continue
			}
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEsc = false
			}
			continue
		}
		out.WriteByte(ch)
	}
	return out.String()
}

func headingAnchor(n *ast.Heading, source []byte, fallback string) string {
	text := extractInlineText(n, source)
	if strings.TrimSpace(text) == "" {
		text = stripANSI(fallback)
	}
	return slugifyAnchor(text)
}

func slugifyAnchor(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func inferLinkKind(url string) LinkKind {
	switch {
	case strings.HasPrefix(url, "#"):
		return LinkKindAnchor
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"), strings.HasPrefix(url, "mailto:"):
		return LinkKindExternal
	default:
		return LinkKindLocal
	}
}

func detectCallout(n *ast.Paragraph, c *renderContext) (string, bool) {
	plain := strings.TrimSpace(extractInlineText(n, c.source))
	firstLine := plain
	if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
		firstLine = firstLine[:idx]
	}
	m := calloutRE.FindStringSubmatch(strings.TrimSpace(firstLine))
	if len(m) != 4 {
		return "", false
	}
	return strings.ToLower(m[1]), true
}

func splitCallout(plain string) (string, string) {
	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) == 0 {
		return "", ""
	}
	m := calloutRE.FindStringSubmatch(strings.TrimSpace(lines[0]))
	title := strings.TrimSpace(lines[0])
	if len(m) == 4 {
		title = strings.TrimSpace(m[3])
		if title == "" {
			title = strings.ToUpper(m[1])
		}
	}
	return title, strings.TrimSpace(strings.Join(lines[1:], "\n"))
}

func collectHTMLText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
		}
		for child := cur.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return b.String()
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

func terminalHyperlink(url, label string) string {
	return "\x1b]8;;" + url + "\x07" + label + "\x1b]8;;\x07"
}

func stripPageBreakMarkup(raw string) string {
	replacements := []string{
		`<div style="page-break-after: always"></div>`,
		`<div style='page-break-after: always'></div>`,
		`<hr class="page-break">`,
		`<hr class='page-break'>`,
	}
	out := raw
	for _, repl := range replacements {
		out = strings.ReplaceAll(out, repl, "")
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func mapRunes(s string, mapping map[rune]rune) string {
	var out strings.Builder
	for _, r := range s {
		if mapped, ok := mapping[r]; ok {
			out.WriteRune(mapped)
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func htmlAnchorHref(raw string) (string, bool) {
	m := htmlAnchorOpenRE.FindStringSubmatch(raw)
	if len(m) == 2 {
		return m[1], true
	}
	return "", false
}

func isHTMLAnchorClose(raw string) bool {
	return strings.EqualFold(strings.TrimSpace(raw), "</a>")
}
