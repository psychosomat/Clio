package app

import (
	"context"
	"strings"
	"time"

	"clio/internal/notes"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := m.update(msg)
	m.syncDimensions()
	return m, cmd
}

func (m *Model) update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 4
		for _, li := range m.Lists {
			li.SetHeight(m.height)
		}
		m.Folders.SetHeight(m.height)
		m.Code.Height = m.height
		m.LineNumbers.Height = m.height
		m.Code.Width = m.width - listWidth - foldersWidth - previewPadding
		m.LineNumbers.Width = 5
		m.syncDimensions()
		return m, nil

	case MsgLoadNotes:
		return m, LoadNotesCmd(m.Store)

	case MsgNotesLoaded:
		m.Notes = msg.Notes
		m.applyFilter()
		m.rebuildFolderLists()
		return m, nil

	case MsgNoteSaved:
		m.CurrentNote = msg.Note
		m.RenameActive = false
		m.RenameInput.Blur()
		if msg.Seq == m.SavingSeq {
			m.Saving = false
		}
		if msg.Seq >= m.AutosaveSeq {
			m.EditorDirty = false
		} else if m.Mode == editorMode {
			m.EditorDirty = true
			return m, AutosaveCmd(m.AutosaveSeq, 150*time.Millisecond)
		}
		return m, LoadNotesCmd(m.Store)

	case MsgNoteCreated:
		m.CurrentNote = msg.Note
		return m, nil

	case MsgAutosaveTick:
		if m.Mode != editorMode || !m.EditorDirty || msg.Seq != m.AutosaveSeq {
			return m, nil
		}
		if m.Saving {
			return m, AutosaveCmd(msg.Seq, 150*time.Millisecond)
		}
		note, ok := m.editorNoteForSave()
		if !ok {
			m.EditorDirty = false
			return m, nil
		}
		m.Saving = true
		m.SavingSeq = msg.Seq
		return m, SaveNoteCmd(m.Store, note, msg.Seq)

	case MsgStatusClear:
		if m.StatusMsgID == msg.ID {
			m.StatusMsg = ""
		}

	case MsgError:
		m.StatusMsg = "Error: " + msg.Err.Error()
		m.StatusIsErr = true

	case MsgLinkOpened:
		m.StatusMsg = "Opened " + msg.Target

	case updateFoldersMsg:
		setItemsCmd := m.Folders.SetItems(msg.items)
		m.Folders.Select(msg.selectedFolderIndex)
		var cmd tea.Cmd
		m.Folders, cmd = m.Folders.Update(msg)
		return m, tea.Batch(setItemsCmd, cmd)

	case updateContentMsg:
		n := Note(msg)
		m.renderPreviewInPane(n.Body)
		return m, nil

	case changeStateMsg:
		if m.List() != nil {
			m.List().SetDelegate(noteDelegate{m.ListStyle, msg.newState})
		}
		if m.browseState == msg.newState {
			break
		}
		wasEditing := m.browseState == navigating
		m.browseState = msg.newState
		m.updateKeyMap()

		switch msg.newState {
		case navigating:
			if wasEditing {
			}
		case copyVisual:
			m.pane = notePane
			m.browseState = copyVisual
			cmd := m.updatePaneStyles(nil)
			return m, tea.Batch(cmd, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return changeStateMsg{navigating}
			}))
		case deleteConfirm:
			m.pane = notePane
			m.browseState = deleteConfirm
			m.updatePaneStyles(nil)
			return m, nil
		}
		return m, nil

	case tea.MouseMsg:
		if m.Mode == browsingMode {
			var cmd tea.Cmd
			switch m.pane {
			case notePane:
				if m.List() != nil {
					*m.List(), cmd = m.List().Update(msg)
				}
			case contentPane:
				m.Code, cmd = m.Code.Update(msg)
				m.LineNumbers, cmd = m.LineNumbers.Update(msg)
			case folderPane:
				m.Folders, cmd = m.Folders.Update(msg)
			}
			return m, cmd
		}
		if m.Mode == editorMode {
			var cmd tea.Cmd
			if m.EditorPreview {
				m.PreviewViewport, cmd = m.PreviewViewport.Update(msg)
				m.LineNumbers, cmd = m.LineNumbers.Update(msg)
			} else {
				m.EditorBody, cmd = m.EditorBody.Update(msg)
			}
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.Mode == helpMode {
			m.Mode = browsingMode
			return m, nil
		}

		if m.Mode == editorMode {
			return m.updateEditorKey(msg)
		}

		if m.RenameActive {
			return m.updateRenameKey(msg)
		}

		if m.FolderActive {
			return m.updateFolderKey(msg)
		}

		if m.browseState == deleteConfirm {
			switch msg.String() {
			case "y":
				item := m.List().SelectedItem()
				if n, ok := item.(Note); ok && n.File != "" {
					m.browseState = navigating
					m.updateKeyMap()
					return m, DeleteNoteCmd(m.Store, n.File)
				}
				m.browseState = navigating
				m.updateKeyMap()
				return m, nil
			case "n", "esc":
				m.browseState = navigating
				m.updateKeyMap()
				return m, nil
			}
			return m, nil
		}

		if m.Mode == searchMode {
			return m.updateSearchKey(msg)
		}

		return m.updateBrowsingKey(msg)
	}

	return m, nil
}

func (m *Model) updateRenameKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		newTitle := strings.TrimSpace(m.RenameInput.Value())
		if newTitle == "" {
			m.RenameActive = false
			m.RenameInput.Blur()
			return m, nil
		}
		item := m.List().SelectedItem()
		if item != nil {
			n := item.(Note)
			nn := m.findNotesNote(n)

			// Update the first line of body to be the new title
			bodyLines := strings.SplitN(nn.Body, "\n", 2)
			if len(bodyLines) > 0 {
				oldFirstLine := strings.TrimSpace(strings.TrimLeft(bodyLines[0], "# "))
				if oldFirstLine != "" && strings.HasPrefix(strings.TrimSpace(bodyLines[0]), "#") {
					bodyLines[0] = "# " + newTitle
					nn.Body = strings.Join(bodyLines, "\n")
				}
			}
			nn.Title = newTitle
			m.Saving = true
			m.SavingSeq = m.AutosaveSeq
			m.RenameActive = false
			m.RenameInput.Blur()
			return m, SaveNoteCmd(m.Store, nn, m.AutosaveSeq)
		}
		m.RenameActive = false
		m.RenameInput.Blur()
		return m, nil
	case tea.KeyEsc:
		m.RenameActive = false
		m.RenameInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.RenameInput, cmd = m.RenameInput.Update(msg)
		return m, cmd
	}
}

func (m *Model) updateFolderKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		folderName := strings.TrimSpace(m.FolderInput.Value())
		if folderName == "" {
			m.FolderActive = false
			m.FolderInput.Blur()
			return m, nil
		}
		item := m.List().SelectedItem()
		if n, ok := item.(Note); ok {
			nn := m.findNotesNote(n)
			nn.Folder = folderName
			m.Saving = true
			m.SavingSeq = m.AutosaveSeq
			m.FolderActive = false
			m.FolderInput.Blur()
			return m, SaveNoteCmd(m.Store, nn, m.AutosaveSeq)
		}
		m.ensureFolderExists(Folder(folderName))
		m.FolderActive = false
		m.FolderInput.Blur()
		return m, nil
	case tea.KeyEsc:
		m.FolderActive = false
		m.FolderInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.FolderInput, cmd = m.FolderInput.Update(msg)
		return m, cmd
	}
}

func (m *Model) updateBrowsingKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.saveState()
		return m, tea.Quit

	case key.Matches(msg, m.keys.ToggleHelp):
		m.Mode = helpMode
		return m, nil

	case key.Matches(msg, m.keys.Search):
		m.Mode = searchMode
		m.SearchInput.SetValue("")
		m.SearchInput.Focus()
		return m, nil

	case key.Matches(msg, m.keys.NewNote):
		m.InitNewNote()
		return m, m.createNewNote()

	case key.Matches(msg, m.keys.OpenNote):
		item := m.List().SelectedItem()
		if item != nil {
			n := item.(Note)
			m.openNote(m.findNotesNote(n))
		}
		return m, nil

	case key.Matches(msg, m.keys.DeleteNote):
		if m.List().SelectedItem() != nil {
			m.pane = notePane
			m.List().Title = "Delete? (y/N)"
			return m, changeState(deleteConfirm)
		}
		return m, nil

	case key.Matches(msg, m.keys.CopyNote):
		return m, changeState(copyVisual)

	case key.Matches(msg, m.keys.EditNote):
		item := m.List().SelectedItem()
		if item != nil {
			n := item.(Note)
			m.openNote(m.findNotesNote(n))
		}
		return m, nil

	case key.Matches(msg, m.keys.Archive):
		return m, m.archiveNote()

	case key.Matches(msg, m.keys.ToggleArchived):
		m.ShowArchived = !m.ShowArchived
		m.applyFilter()
		m.rebuildFolderLists()
		if m.ShowArchived {
			m.StatusMsg = "Showing archived"
		} else {
			m.StatusMsg = "Hiding archived"
		}
		m.StatusIsErr = false
		return m, nil

	case key.Matches(msg, m.keys.RenameNote):
		item := m.List().SelectedItem()
		if item != nil {
			n := item.(Note)
			m.RenameInput.SetValue(n.Title)
			m.RenameActive = true
			m.RenameInput.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.SetFolder):
		item := m.List().SelectedItem()
		if item != nil {
			n := item.(Note)
			m.FolderInput.SetValue(n.Folder)
			m.FolderActive = true
			m.FolderInput.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.CreateFolder):
		return m, m.createFolder()

	case key.Matches(msg, m.keys.NextPane):
		m.nextPane()
		_ = m.updatePaneStyles(msg)
		if m.pane == contentPane {
			item := m.List().SelectedItem()
			if item != nil {
				m.renderPreviewInPane(item.(Note).Body)
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.PreviousPane):
		m.previousPane()
		_ = m.updatePaneStyles(msg)
		if m.pane == contentPane {
			item := m.List().SelectedItem()
			if item != nil {
				m.renderPreviewInPane(item.(Note).Body)
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.MoveNoteUp):
		if m.pane == notePane {
			return m, m.moveNoteUp()
		}

	case key.Matches(msg, m.keys.MoveNoteDown):
		if m.pane == notePane {
			return m, m.moveNoteDown()
		}
	}

	cmd := m.updatePaneStyles(msg)
	return m, cmd
}

func (m *Model) updateSearchKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.SearchInput.Value() != "" {
			m.SearchInput.SetValue("")
			m.applyFilter()
			m.rebuildFolderLists()
		} else {
			m.SearchInput.Blur()
			m.Mode = browsingMode
			m.rebuildFolderLists()
		}
		return m, nil
	case tea.KeyEnter:
		m.SearchInput.Blur()
		m.Mode = browsingMode
		m.rebuildFolderLists()
		return m, nil
	default:
		var cmd tea.Cmd
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.applyFilter()
		m.rebuildFolderLists()
		return m, cmd
	}
}

func (m *Model) updateEditorKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	var cmd tea.Cmd

	if key.Matches(msg, m.keys.Esc) {
		m.Mode = browsingMode
		if m.EditorDirty && !m.Saving {
			note, ok := m.editorNoteForSave()
			if ok {
				m.Saving = true
				m.SavingSeq = m.AutosaveSeq
				return m, SaveNoteCmd(m.Store, note, m.AutosaveSeq)
			}
		}
		if m.CurrentNote.ID != "" && strings.TrimSpace(m.EditorBody.Value()) == "" {
			return m, DeleteNoteCmd(m.Store, m.CurrentNote.ID)
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Preview) {
		m.EditorPreview = !m.EditorPreview
		if m.EditorPreview {
			previewHeight := m.height - 8
			if previewHeight < 5 {
				previewHeight = 5
			}
			m.PreviewViewport.Height = previewHeight
			m.LineNumbers.Height = previewHeight
			m.LineNumbers.Width = 5
			m.updateMarkdownPreview()
		}
		return m, nil
	}

	if m.EditorPreview {
		switch msg.String() {
		case "tab":
			m.cyclePreviewLink(1)
			m.updateMarkdownPreview()
			return m, nil
		case "shift+tab":
			m.cyclePreviewLink(-1)
			m.updateMarkdownPreview()
			return m, nil
		case "enter":
			return m, m.openPreviewLink()
		}
	}

	if key.Matches(msg, m.keys.EditExternal) {
		return m, m.editExternally()
	}

	if snippet, ok := snippetForKey(msg.String()); ok {
		line := m.EditorBody.Line()
		column := m.EditorBody.LineInfo().CharOffset
		updated, cursorLine, cursorCol := insertSnippet(m.EditorBody.Value(), line, column, snippet.body)
		m.EditorBody.SetValue(updated)
		m.restoreEditorCursor(cursorLine, cursorCol)
		m.EditorDirty = true
		m.AutosaveSeq++
		return m, AutosaveCmd(m.AutosaveSeq, 700*time.Millisecond)
	}

	oldVal := m.EditorBody.Value()
	m.EditorBody, cmd = m.EditorBody.Update(msg)
	if m.EditorBody.Value() != oldVal {
		m.EditorDirty = true
		m.AutosaveSeq++
		return m, tea.Batch(cmd, AutosaveCmd(m.AutosaveSeq, 700*time.Millisecond))
	}

	return m, cmd
}

func (m *Model) syncDimensions() {
	if m.width == 0 || m.height == 0 {
		return
	}

	editorBodyHeight := m.height - 8
	if editorBodyHeight < 5 {
		editorBodyHeight = 5
	}
	m.EditorBody.SetWidth(m.width - 4)
	m.EditorBody.SetHeight(editorBodyHeight)

	if m.Mode == editorMode {
		previewWidth := m.width - 11
		if previewWidth < 20 {
			previewWidth = 20
		}
		m.PreviewViewport.Width = previewWidth
		m.PreviewViewport.Height = editorBodyHeight
		m.LineNumbers.Height = editorBodyHeight
		m.LineNumbers.Width = 5
	}
}

func (m *Model) applyFilter() {
	query := m.SearchInput.Value()
	var filtered []notes.Note
	for _, n := range m.Notes {
		if n.Archived && !m.ShowArchived {
			continue
		}
		if query != "" {
			q := strings.ToLower(query)
			title := strings.ToLower(n.DisplayTitle())
			body := strings.ToLower(n.Body)
			if !strings.Contains(title, q) && !strings.Contains(body, q) {
				continue
			}
		}
		filtered = append(filtered, n)
	}
	m.Filtered = filtered
}

func (m *Model) findNotesNote(n Note) notes.Note {
	for _, nn := range m.Filtered {
		if nn.ID == n.ID {
			return nn
		}
		if nn.DisplayTitle() == n.Title {
			return nn
		}
	}
	return m.CurrentNote
}

func (m *Model) archiveNote() tea.Cmd {
	item := m.List().SelectedItem()
	if item == nil {
		return nil
	}
	n := m.findNotesNote(item.(Note))
	n.Archived = !n.Archived
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.Store.Save(ctx, n)
	if err != nil {
		return nil
	}
	return LoadNotesCmd(m.Store)
}

func (m *Model) editExternally() tea.Cmd {
	note, ok := m.editorNoteForSave()
	if !ok {
		return nil
	}
	m.Saving = true
	m.SavingSeq = m.AutosaveSeq
	path := note.Path
	if path == "" {
		path = m.Store.(*notes.FileStore).NotesDir + "/" + note.ID + ".md"
	}
	return tea.Batch(
		SaveNoteCmd(m.Store, note, m.AutosaveSeq),
		tea.ExecProcess(editorCmd(path), func(err error) tea.Msg {
			if err != nil {
				return MsgError{Err: err}
			}
			return MsgLoadNotes{}
		}),
	)
}
