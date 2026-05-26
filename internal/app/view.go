package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing Clio..."
	}

	var content string

	switch m.Mode {
	case browsingMode:
		content = m.renderBrowsing()
	case editorMode:
		content = m.renderEditor()
	case searchMode:
		content = m.renderSearch()
	case helpMode:
		content = m.renderHelp()
	}

	if m.RenameActive {
		overlay := m.renderInputOverlay(m.RenameInput.View(), "Rename Note")
		content = lipgloss.Place(m.width, m.height+4, lipgloss.Center, lipgloss.Center, overlay)
	}

	if m.FolderActive {
		overlay := m.renderInputOverlay(m.FolderInput.View(), "Folder")
		content = lipgloss.Place(m.width, m.height+4, lipgloss.Center, lipgloss.Center, overlay)
	}

	if m.browseState == deleteConfirm {
		note := m.selectedNote()
		overlay := m.renderDeleteConfirmOverlay(note.Title)
		content = lipgloss.Place(m.width, m.height+4, lipgloss.Center, lipgloss.Center, overlay)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height + 4).
		Render(content)
}

func (m *Model) renderDeleteConfirmOverlay(noteTitle string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AlertColor).
		Padding(1, 2).
		Width(50)

	title := lipgloss.NewStyle().Foreground(AlertColor).Bold(true).Render("Delete Note?")
	body := fmt.Sprintf("\"%s\"\n\n%s  %s",
		truncateString(noteTitle, 40),
		lipgloss.NewStyle().Foreground(PrimaryColor).Render("[y]"),
		MutedStyle.Render("delete  •  [n] cancel"),
	)
	return box.Render(title + "\n\n" + body)
}

func (m *Model) renderInputOverlay(inputView, title string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(40)

	content := lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true).Render(title) + "\n\n" + inputView
	return box.Render(content)
}

func (m *Model) renderBrowsing() string {
	var sb strings.Builder

	modeText := "CLIO"
	countStr := fmt.Sprintf("%d notes", len(m.Filtered))
	headerStr := TitleStyle.Render(" " + modeText + " ")
	padding := m.width - lipgloss.Width(headerStr) - lipgloss.Width(countStr) - 2
	if padding > 0 {
		sb.WriteString(headerStr + strings.Repeat(" ", padding) + countStr)
	} else {
		sb.WriteString(headerStr + " " + countStr)
	}
	sb.WriteString("\n\n")

	availableHeight := m.height
	if availableHeight < 5 {
		availableHeight = 5
	}

	defStyles := DefaultStyles(m.config)
	notesStyle := defStyles.Notes.Blurred
	previewStyle := defStyles.Content.Blurred
	foldersStyle := defStyles.Folders.Blurred

	switch m.pane {
	case notePane:
		notesStyle = m.ListStyle
	case contentPane:
		previewStyle = m.ContentStyle
	case folderPane:
		foldersStyle = m.FoldersStyle
	}

	var mainBody string
	items := m.List().Items()
	if len(items) == 0 {
		placeholder := "\n\n  No notes found.\n  Press n to create a new note."
		mainBody = lipgloss.NewStyle().Height(availableHeight).Render(placeholder)
	} else {
		titleBar := notesStyle.TitleBar.Render("Notes")
		var listSb strings.Builder
		listSb.WriteString(titleBar)

		for i, item := range items {
			n, ok := item.(Note)
			if !ok {
				continue
			}
			listSb.WriteString(m.renderNoteRow(n, i == m.List().Index(), notesStyle))
			listSb.WriteString("\n")
		}
		listPane := notesStyle.Base.
			Height(availableHeight).
			Width(38).
			Render(listSb.String())

		sel := m.selectedNote()
		previewTitle := previewStyle.Title.Render(" " + sel.Title + " ")
		previewWidth := m.width - listWidth - foldersWidth - lineNumWidth - 10
		if previewWidth < 20 {
			previewWidth = 20
		}

		// Limit preview content to available height to prevent artifacts
		previewContent := previewStyle.Base.
			Height(availableHeight - 2).
			Width(previewWidth + 5).
			Render(
				lipgloss.JoinHorizontal(lipgloss.Top,
					previewStyle.LineNumber.
						Height(availableHeight-2).
						Render(m.LineNumbers.View()),
					lipgloss.NewStyle().
						Width(previewWidth).
						MaxWidth(previewWidth).
						Render(m.Code.View()),
				),
			)

		previewPane := lipgloss.NewStyle().
			Height(availableHeight).
			Width(previewWidth + 5).
			Render(previewTitle + previewContent)

		folderTitle := foldersStyle.TitleBar.Render("Folders")
		var folderSb strings.Builder
		folderSb.WriteString(folderTitle)
		folderSb.WriteString(m.Folders.View())
		folderPane := foldersStyle.Base.
			Height(availableHeight).
			Width(22).
			Render(folderSb.String())

		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, folderPane, listPane, "  ", previewPane)
	}
	sb.WriteString(mainBody)
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m *Model) renderSearch() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" CLIO ") + " " + ModeStyle.Render("SEARCH"))
	sb.WriteString("\n\n")
	sb.WriteString(m.SearchInput.View())
	sb.WriteString("\n\n")

	for _, n := range m.Filtered {
		sb.WriteString("  " + MutedStyle.Render(n.DisplayTitle()) + "\n")
	}
	sb.WriteString(m.renderFooter())
	return sb.String()
}

func (m *Model) renderEditor() string {
	var sb strings.Builder

	headerText := " CLIO "
	modeText := "NEW NOTE"
	if m.CurrentNote.ID != "" {
		modeText = "EDIT NOTE"
	}
	if m.EditorPreview {
		modeText += " (PREVIEW)"
	}
	sb.WriteString(TitleStyle.Render(headerText) + " " + ModeStyle.Render(modeText) + "\n\n")

	title := m.CurrentNote.DisplayTitle()
	sb.WriteString(MutedStyle.Render(" "+title) + "\n\n")

	editorHeight := m.height - 8
	if editorHeight < 5 {
		editorHeight = 5
	}

	if m.EditorPreview {
		previewContent := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(5).Render(m.LineNumbers.View()),
			m.PreviewViewport.View(),
		)
		sb.WriteString(ActiveBorderedBox.
			Width(m.width - 2).
			Height(editorHeight).
			Render(previewContent))
	} else {
		sb.WriteString(ActiveBorderedBox.
			Width(m.width - 2).
			Height(editorHeight).
			Render(m.EditorBody.View()))
	}

	sb.WriteString(m.renderFooter())
	return sb.String()
}

func (m *Model) renderHelp() string {
	var sb strings.Builder
	sb.WriteString(AccentTitleStyle.Render(" Clio Shortcuts ") + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Foreground(PrimaryColor).Render("Quick Keys:") + "\n")
	for _, b := range m.keys.ShortHelp() {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			lipgloss.NewStyle().Foreground(SecondaryColor).Render(b.Help().Key),
			b.Help().Desc,
		))
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Foreground(PrimaryColor).Render("All Commands:") + "\n")
	for _, group := range m.keys.FullHelp() {
		for _, b := range group {
			sb.WriteString(fmt.Sprintf("  %s  %s\n",
				lipgloss.NewStyle().Foreground(SecondaryColor).Render(b.Help().Key),
				b.Help().Desc,
			))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(HelpStyle.Render("Default store: ~/.local/share/clio/notes") + "\n")

	boxed := ModalBox.Render(sb.String())
	return lipgloss.Place(m.width, m.height+4, lipgloss.Center, lipgloss.Center, boxed)
}

func (m *Model) renderNoteRow(note Note, isSelected bool, style NotesBaseStyle) string {
	title := strings.ReplaceAll(note.Title, "\n", " ")
	title = strings.TrimSpace(title)

	timeStr := relativeTime(note.Date)
	width := 35
	leftWidth := width - len(timeStr) - 3
	if leftWidth < 10 {
		leftWidth = 10
	}
	title = truncateString(title, leftWidth)

	var mainStr string
	if isSelected {
		mainStr = style.SelectedTitle.Render(title)
	} else {
		mainStr = style.UnselectedTitle.Render(title)
	}

	rightStr := style.UnselectedSubtitle.Render(timeStr)
	rowPadding := width - lipgloss.Width(mainStr) - lipgloss.Width(rightStr) - 2
	if rowPadding < 0 {
		rowPadding = 0
	}
	paddingStr := strings.Repeat(" ", rowPadding)

	return fmt.Sprintf("  %s%s%s", mainStr, paddingStr, rightStr)
}

func (m *Model) renderFooter() string {
	var sb strings.Builder

	sb.WriteString(MutedStyle.Render(strings.Repeat("─", m.width)) + "\n")

	status := ""
	if m.StatusMsg != "" {
		if m.StatusIsErr {
			status = AlertStyle.Render(m.StatusMsg)
		} else {
			status = SuccessStyle.Render(m.StatusMsg)
		}
	}

	var hints []string
	switch m.Mode {
	case browsingMode:
		if m.pane == notePane {
			hints = []string{"tab focus", "↑/k up", "↓/j down", "n new", "enter open", "/ search", "? help", "q quit"}
		} else {
			hints = []string{"tab focus", "↑ scroll", "↓ scroll", "? help"}
		}
	case editorMode:
		if m.EditorPreview {
			hints = []string{"f2 edit", "tab link", "enter open", "esc back"}
		} else {
			hints = []string{"e external", "f2 preview", "f3 code", "esc back"}
		}
	case searchMode:
		hints = []string{"type to filter", "esc clear", "enter done"}
	case helpMode:
		hints = []string{"any key to close"}
	}

	hintsStr := MutedStyle.Render(strings.Join(hints, "  •  "))

	if status != "" {
		padding := m.width - lipgloss.Width(status) - lipgloss.Width(hintsStr) - 2
		if padding > 0 {
			sb.WriteString(" " + status + strings.Repeat(" ", padding) + hintsStr + " ")
		} else {
			sb.WriteString(" " + status + " | " + hintsStr)
		}
	} else {
		sb.WriteString(" " + hintsStr)
	}

	return sb.String()
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	if d < 5*time.Second {
		return "now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return t.Format("Jan 2")
}

func truncateString(s string, limit int) string {
	if lipgloss.Width(s) <= limit {
		return s
	}
	var count int
	var sb strings.Builder
	for _, r := range s {
		sb.WriteRune(r)
		count++
		if count >= limit-3 {
			sb.WriteString("...")
			break
		}
	}
	return sb.String()
}
