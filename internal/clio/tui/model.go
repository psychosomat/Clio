package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	appconfig "clio/internal/clio/config"
	appdomain "clio/internal/clio/domain"
	appservice "clio/internal/clio/service"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const maxPane = 2

type pane int

const (
	notePane pane = iota
	folderPane
	contentPane
)

type state int

const (
	navigatingState state = iota
	deletingState
	creatingState
	copyingState
	pastingState
	quittingState
	editingState
	editingTagsState
	contentEditingState
	creatingFolderState
	deletingFolderState
)

type input int

const (
	folderInput input = iota
	nameInput
	languageInput
)

const (
	defaultNoteFolder = appdomain.DefaultNotebook
	defaultNoteName   = appdomain.UntitledNoteTitle
)

type Model struct {
	config      appconfig.App
	keys        KeyMap
	help        help.Model
	height      int
	showPreview bool
	Workdir     string

	Lists        map[Folder]*list.Model
	Folders      list.Model
	Code         viewport.Model
	LineNumbers  viewport.Model
	Textarea     textarea.Model
	activeInput  input
	inputs       []textinput.Model
	tagsInput        textinput.Model
	createFolderInput textinput.Model
	pane             pane
	state        state
	ListStyle    NotesBaseStyle
	FoldersStyle FoldersBaseStyle
	ContentStyle ContentBaseStyle
	library      *appservice.Service
}

func (m *Model) Init() tea.Cmd {
	rand.Seed(time.Now().Unix())

	m.Folders.Styles.Title = m.FoldersStyle.Title
	m.Folders.Styles.TitleBar = m.FoldersStyle.TitleBar
	m.updateKeyMap()

	return func() tea.Msg {
		return updateContentMsg{note: m.selectedNote()}
	}
}

type updateContentMsg struct {
	note appdomain.Note
}

func (m *Model) updateContent() tea.Cmd {
	return func() tea.Msg {
		return updateContentMsg{note: m.selectedNote()}
	}
}

type updateFoldersMsg struct {
	items               []list.Item
	selectedFolderIndex int
}

func (m *Model) updateFolders() tea.Cmd {
	return func() tea.Msg {
		msg := m.updateFoldersView()
		return msg
	}
}

type changeStateMsg struct{ newState state }

func changeState(newState state) tea.Cmd {
	return func() tea.Msg {
		return changeStateMsg{newState}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == contentEditingState {
		switch msg := msg.(type) {
		case changeStateMsg:
		case tea.WindowSizeMsg:
			m.height = msg.Height - 4
			for _, li := range m.Lists {
				li.SetHeight(m.height)
			}
			m.Folders.SetHeight(m.height)
			m.Code.Height = m.height
			m.LineNumbers.Height = m.height
			m.Code.Width = msg.Width - m.List().Width() - m.Folders.Width() - 20
			m.LineNumbers.Width = 5
			m.Textarea.SetWidth(m.Code.Width - 4)
			m.Textarea.SetHeight(m.height)
			var cmd tea.Cmd
			m.Textarea, cmd = m.Textarea.Update(msg)
			return m, cmd
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.Cancel):
				return m, changeState(navigatingState)
			case key.Matches(msg, m.keys.TogglePreview):
				return m, m.togglePreview()
			case key.Matches(msg, m.keys.InsertCodeBlock):
				m.Textarea.InsertString("\n```text\n// your code here\n```\n")
				return m, m.refreshPreview()
			case key.Matches(msg, m.keys.InsertTable):
				m.Textarea.InsertString("\n| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |\n")
				return m, m.refreshPreview()
			case key.Matches(msg, m.keys.InsertChecklist):
				m.Textarea.InsertString("\n- [ ] task\n")
				return m, m.refreshPreview()
			case key.Matches(msg, m.keys.InsertQuote):
				m.Textarea.InsertString("\n> quote\n")
				return m, m.refreshPreview()
			case key.Matches(msg, m.keys.InsertLink):
				m.Textarea.InsertString("\n[link title](url)\n")
				return m, m.refreshPreview()
			}
			if m.showPreview {
				return m, nil
			}
			var cmd tea.Cmd
			m.Textarea, cmd = m.Textarea.Update(msg)
			return m, cmd
		default:
			var cmd tea.Cmd
			m.Textarea, cmd = m.Textarea.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case updateFoldersMsg:
		setItemsCmd := m.Folders.SetItems(msg.items)
		m.Folders.Select(msg.selectedFolderIndex)
		var cmd tea.Cmd
		m.Folders, cmd = m.Folders.Update(msg)
		return m, tea.Batch(setItemsCmd, cmd)
	case updateContentMsg:
		return m.updateContentView(msg)
	case changeStateMsg:
		m.List().SetDelegate(noteDelegate{m.ListStyle, msg.newState})

		var cmd tea.Cmd

		if m.state == msg.newState {
			break
		}

		wasEditing := m.state == editingState
		wasPasting := m.state == pastingState
		wasCreating := m.state == creatingState
		wasContentEditing := m.state == contentEditingState
		wasDeletingFolder := m.state == deletingFolderState
		m.state = msg.newState
		m.updateKeyMap()
		m.updateActivePane(msg)

		switch msg.newState {
		case navigatingState:
			if wasContentEditing {
				m.showPreview = true
				content := m.Textarea.Value()
				if err := m.library.Repository.WriteContent(m.selectedNote(), content); err != nil {
					return m, changeState(navigatingState)
				}
				cmd = m.updateContent()
			}

			if wasPasting || wasCreating {
				return m, m.updateContent()
			}

			m.Folders.Title = "Notebooks"

			if wasDeletingFolder {
				for i, item := range m.Folders.Items() {
					if item.FilterValue() == appdomain.DefaultNotebook {
						m.Folders.Select(i)
						break
					}
				}
				cmd = tea.Batch(m.updateContent())
			}

			if wasEditing {
				m.blurInputs()
				i := m.List().Index()
				note := m.selectedNote()
				if m.inputs[nameInput].Value() != "" {
					note.Title = m.inputs[nameInput].Value()
				} else {
					note.Title = defaultNoteName
				}
				if m.inputs[folderInput].Value() != "" {
					note.Notebook = m.inputs[folderInput].Value()
				} else {
					note.Notebook = defaultNoteFolder
				}
				if m.inputs[languageInput].Value() != "" {
					note.Format = m.inputs[languageInput].Value()
				} else {
					note.Format = ""
				}
				updated, err := m.library.UpdateNote(m.selectedNote(), note)
				if err == nil {
					note = updated
				}
				setCmd := m.List().SetItem(i, NoteItem{note})
				m.pane = notePane
				cmd = tea.Batch(setCmd, m.updateFolders(), m.updateContent())
			}
		case contentEditingState:
			m.pane = contentPane
			m.showPreview = false
			content, err := m.library.Repository.ReadContent(m.selectedNote())
			if err != nil {
				content = ""
			}
			m.Textarea.SetValue(content)
			m.Textarea.Focus()
			cmd = textarea.Blink
		case pastingState:
			content, err := clipboard.ReadAll()
			if err != nil {
				return m, changeState(navigatingState)
			}
			if err := m.library.AppendToNote(m.selectedNote(), content); err != nil {
				return m, changeState(navigatingState)
			}
			return m, changeState(navigatingState)
		case deletingState:
			m.state = deletingState
		case editingState:
			m.pane = contentPane
			note := m.selectedNote()
			m.inputs[folderInput].SetValue(note.Notebook)
			if note.Title == defaultNoteName {
				m.inputs[nameInput].SetValue("")
			} else {
				m.inputs[nameInput].SetValue(note.Title)
			}
			m.inputs[languageInput].SetValue(note.Format)
			cmd = m.focusInput(m.activeInput)
		case creatingFolderState:
			m.state = creatingFolderState
			m.pane = folderPane
		case deletingFolderState:
			m.state = deletingFolderState
			m.pane = folderPane
		case creatingState:
		case copyingState:
			m.pane = notePane
			m.state = copyingState
			m.updateActivePane(msg)
			cmd = tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return changeStateMsg{navigatingState}
			})
		}

		m.updateKeyMap()
		m.updateActivePane(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.height = msg.Height - 4
		for _, li := range m.Lists {
			li.SetHeight(m.height)
		}
		m.Folders.SetHeight(m.height)
		m.Code.Height = m.height
		m.LineNumbers.Height = m.height
		m.Code.Width = msg.Width - m.List().Width() - m.Folders.Width() - 20
		m.LineNumbers.Width = 5
		m.Textarea.SetWidth(m.Code.Width - 4)
		m.Textarea.SetHeight(m.height)
		return m, nil
	case tea.KeyMsg:
		if m.List().FilterState() == list.Filtering {
			break
		}

		if m.state == deletingState {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				_ = m.library.DeleteNote(m.selectedNote())
				m.List().RemoveItem(m.List().Index())
				m.state = navigatingState
				m.updateKeyMap()
				return m, tea.Batch(changeState(navigatingState), func() tea.Msg {
					return updateContentMsg{note: m.selectedNote()}
				})
			case key.Matches(msg, m.keys.Quit, m.keys.Cancel):
				return m, changeState(navigatingState)
			}
			return m, nil
		} else if m.state == copyingState {
			return m, changeState(navigatingState)
		} else if m.state == deletingFolderState {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				m.deleteFolder()
				return m, changeState(navigatingState)
			case key.Matches(msg, m.keys.Quit, m.keys.Cancel):
				return m, changeState(navigatingState)
			}
			return m, nil
		} else if m.state == editingState {
			if msg.String() == "esc" || msg.String() == "enter" {
				return m, changeState(navigatingState)
			}
			var cmd tea.Cmd
			var cmds []tea.Cmd
			for i := range m.inputs {
				m.inputs[i], cmd = m.inputs[i].Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		} else if m.state == creatingFolderState {
			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.createFolderInput.Blur()
				return m, changeState(navigatingState)
			case msg.String() == "enter":
				name := m.createFolderInput.Value()
				name = strings.TrimSpace(name)
				if name == "" {
					return m, nil
				}
				if err := m.library.CreateFolder(name); err != nil {
					return m, nil
				}
				m.Lists[Folder(name)] = newList([]list.Item{}, m.height, m.ListStyle)
				m.createFolderInput.Blur()
				var folderItems []list.Item
				foldersSlice := maps.Keys(m.Lists)
				slices.Sort(foldersSlice)
				for _, f := range foldersSlice {
					folderItems = append(folderItems, Folder(f))
				}
				m.Folders.SetItems(folderItems)
				for i, f := range foldersSlice {
					if string(f) == name {
						m.Folders.Select(i)
						break
					}
				}
				return m, changeState(navigatingState)
			default:
				var cmd tea.Cmd
				m.createFolderInput, cmd = m.createFolderInput.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.NextPane):
			m.nextPane()
		case key.Matches(msg, m.keys.PreviousPane):
			m.previousPane()
		case key.Matches(msg, m.keys.Quit):
			m.saveState()
			m.state = quittingState
			return m, tea.Quit
		case key.Matches(msg, m.keys.NewNote):
			m.state = creatingState
			return m, m.createNewNoteFile()
		case key.Matches(msg, m.keys.NewFolder):
			m.createFolderInput.SetValue("")
			m.createFolderInput.Focus()
			m.pane = folderPane
			return m, changeState(creatingFolderState)
		case key.Matches(msg, m.keys.DeleteFolder):
			if m.selectedFolder() == Folder(appdomain.DefaultNotebook) {
				return m, nil
			}
			m.pane = folderPane
			return m, changeState(deletingFolderState)
		case key.Matches(msg, m.keys.MoveNoteDown):
			m.moveNoteDown()
		case key.Matches(msg, m.keys.MoveNoteUp):
			m.moveNoteUp()
		case key.Matches(msg, m.keys.PasteNote):
			return m, changeState(pastingState)
		case key.Matches(msg, m.keys.RenameNote):
			m.activeInput = nameInput
			return m, changeState(editingState)
		case key.Matches(msg, m.keys.ChangeFolder):
			m.pane = notePane
			cmd := m.updateActivePane(msg)
			return m, cmd
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll

			var newHeight int
			if m.help.ShowAll {
				newHeight = m.height - 4
			} else {
				newHeight = m.height
			}
			m.List().SetHeight(newHeight)
			m.Folders.SetHeight(newHeight)
			m.Code.Height = newHeight
			m.LineNumbers.Height = newHeight
		case key.Matches(msg, m.keys.TogglePreview):
			m.showPreview = !m.showPreview
			return m, m.updateContent()
		case key.Matches(msg, m.keys.SetFolder):
			m.activeInput = folderInput
			return m, changeState(editingState)
		case key.Matches(msg, m.keys.SetLanguage):
			m.activeInput = languageInput
			return m, changeState(editingState)
		case key.Matches(msg, m.keys.CopyNote):
			return m, func() tea.Msg {
				content, err := m.library.Repository.ReadContent(m.selectedNote())
				if err != nil {
					return changeStateMsg{navigatingState}
				}
				clipboard.WriteAll(content)
				return changeStateMsg{copyingState}
			}
		case key.Matches(msg, m.keys.DeleteNote):
			m.pane = notePane
			m.updateActivePane(msg)
			m.List().Title = "Delete? (y/N)"
			return m, changeState(deletingState)
		case key.Matches(msg, m.keys.EditNote):
			return m, m.editNote()
		case key.Matches(msg, m.keys.Search):
			m.pane = notePane
		}
	}

	m.updateKeyMap()
	cmd := m.updateActivePane(msg)
	return m, cmd
}

func (m *Model) blurInputs() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func (m *Model) focusInput(i input) tea.Cmd {
	m.blurInputs()
	m.inputs[i].CursorEnd()
	return m.inputs[i].Focus()
}

func (m *Model) selectedNoteFilePath() string {
	return m.library.NotePath(m.selectedNote())
}

func (m *Model) nextPane() {
	m.pane = (m.pane + 1) % maxPane
}

func (m *Model) previousPane() {
	m.pane--
	if m.pane < 0 {
		m.pane = maxPane - 1
	}
}

func (m *Model) editNote() tea.Cmd {
	return changeState(contentEditingState)
}

func (m *Model) noContentHints() []keyHint {
	return []keyHint{
		{m.keys.EditNote, "edit contents"},
		{m.keys.PasteNote, "paste clipboard"},
		{m.keys.RenameNote, "rename"},
		{m.keys.SetFolder, "move notebook"},
		{m.keys.SetLanguage, "set format"},
	}
}

func (m *Model) editingContentHints() []keyHint {
	return []keyHint{
		{m.keys.Cancel, "save & close editor"},
	}
}

func (m *Model) updateFoldersView() tea.Msg {
	var selectedFolder Folder
	selectedFolderIndex := m.Folders.Index()
	for folder, li := range m.Lists {
		for i, item := range li.Items() {
			noteItem, ok := item.(NoteItem)
			if !ok {
				continue
			}
			f := Folder(noteItem.Note.Notebook)
			_, ok = m.Lists[f]
			if !ok {
				m.Lists[f] = newList([]list.Item{}, m.height, m.ListStyle)
				selectedFolder = f
			}
			if f != folder {
				li.RemoveItem(i)
				m.Lists[f].InsertItem(0, item)
				selectedFolder = f
			}
		}
	}
	var folderItems []list.Item

	foldersSlice := maps.Keys(m.Lists)
	slices.Sort(foldersSlice)
	for i, folder := range foldersSlice {
		folderItems = append(folderItems, Folder(folder))
		if folder == selectedFolder {
			selectedFolderIndex = i
		}
	}

	return updateFoldersMsg{
		items:               folderItems,
		selectedFolderIndex: selectedFolderIndex,
	}
}

func (m *Model) refreshPreview() tea.Cmd {
	if !m.showPreview {
		return nil
	}
	content := m.Textarea.Value()
	if content == "" {
		return nil
	}
	s := m.library.RenderMarkdown(content)
	if s == "" {
		s = content
	}
	m.writeLineNumbers(lipgloss.Height(s))
	m.Code.SetContent(s)
	return nil
}

func (m *Model) togglePreview() tea.Cmd {
	m.showPreview = !m.showPreview
	if m.showPreview {
		m.refreshPreview()
	}
	m.updateKeyMap()
	return nil
}

func (m *Model) updateContentView(msg updateContentMsg) (tea.Model, tea.Cmd) {
	if len(m.List().Items()) <= 0 {
		m.displayKeyHint([]keyHint{
			{m.keys.NewNote, "create a new note."},
		})
		return m, nil
	}

	content, err := m.library.Repository.ReadContent(msg.note)
	if err != nil {
		m.displayKeyHint(m.noContentHints())
		return m, nil
	}

	if content == "" {
		m.displayKeyHint(m.noContentHints())
		return m, nil
	}

	if m.showPreview {
		s := m.library.RenderContent(msg.note, true)
		if s == "" {
			s = content
		}
		m.writeLineNumbers(lipgloss.Height(s))
		m.Code.SetContent(s)
	} else {
		m.writeLineNumbers(lipgloss.Height(content))
		m.Code.SetContent(content)
	}
	return m, nil
}

type keyHint struct {
	binding key.Binding
	help    string
}

func (m *Model) displayKeyHint(hints []keyHint) {
	m.LineNumbers.SetContent(strings.Repeat("  ~ \n", len(hints)))
	var s strings.Builder
	for _, hint := range hints {
		s.WriteString(
			fmt.Sprintf("%s %s\n",
				m.ContentStyle.EmptyHintKey.Render(hint.binding.Help().Key),
				m.ContentStyle.EmptyHint.Render("• "+hint.help),
			))
	}
	m.Code.SetContent(s.String())
}

func (m *Model) displayError(error string) {
	m.LineNumbers.SetContent(" ~ ")
	m.Code.SetContent(fmt.Sprintf("%s",
		m.ContentStyle.EmptyHint.Render(error),
	))
}

func (m *Model) writeLineNumbers(n int) {
	var lineNumbers strings.Builder
	for i := 1; i < n; i++ {
		lineNumbers.WriteString(fmt.Sprintf("%3d \n", i))
	}
	m.LineNumbers.SetContent(lineNumbers.String() + "  ~ \n")
}

const tabSpaces = 4

func (m *Model) updateActivePane(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch m.pane {
	case folderPane:
		m.ListStyle = DefaultStyles(m.config).Notes.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
		m.FoldersStyle = DefaultStyles(m.config).Folders.Focused
		m.Folders, cmd = m.Folders.Update(msg)
		m.updateKeyMap()
		cmds = append(cmds, cmd, m.updateContent())
	case notePane:
		m.ListStyle = DefaultStyles(m.config).Notes.Focused
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
		m.FoldersStyle = DefaultStyles(m.config).Folders.Blurred
		*m.List(), cmd = (*m.List()).Update(msg)
		cmds = append(cmds, cmd)
	case contentPane:
		m.ListStyle = DefaultStyles(m.config).Notes.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Focused
		m.FoldersStyle = DefaultStyles(m.config).Folders.Blurred
		m.Code, cmd = m.Code.Update(msg)
		cmds = append(cmds, cmd)
		m.LineNumbers, cmd = m.LineNumbers.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.List().SetDelegate(noteDelegate{m.ListStyle, m.state})
	m.Folders.SetDelegate(folderDelegate{m.FoldersStyle})
	m.Folders.Styles.TitleBar = m.FoldersStyle.TitleBar
	m.Folders.Styles.Title = m.FoldersStyle.Title

	return tea.Batch(cmds...)
}

func (m *Model) updateKeyMap() {
	hasItems := len(m.List().VisibleItems()) > 0
	isFiltering := m.List().FilterState() == list.Filtering
	isEditing := m.state == editingState
	isContentEditing := m.state == contentEditingState
	isCreatingFolder := m.state == creatingFolderState
	isDeletingFolder := m.state == deletingFolderState
	isBusy := isFiltering || isEditing || isCreatingFolder || isDeletingFolder
	m.keys.DeleteNote.SetEnabled(hasItems && !isBusy)
	m.keys.CopyNote.SetEnabled(hasItems && !isBusy)
	m.keys.PasteNote.SetEnabled(hasItems && !isBusy)
	m.keys.EditNote.SetEnabled(hasItems && !isBusy)
	m.keys.NewNote.SetEnabled(!isBusy)
	m.keys.NewFolder.SetEnabled(!isBusy)
	m.keys.DeleteFolder.SetEnabled(!isBusy)
	m.keys.ChangeFolder.SetEnabled(m.pane == folderPane)
	m.keys.TogglePreview.SetEnabled(!isCreatingFolder && !isDeletingFolder && !isEditing)
	m.keys.InsertCodeBlock.SetEnabled(isContentEditing)
	m.keys.InsertTable.SetEnabled(isContentEditing)
	m.keys.InsertChecklist.SetEnabled(isContentEditing)
	m.keys.InsertQuote.SetEnabled(isContentEditing)
	m.keys.InsertLink.SetEnabled(isContentEditing)
}

func (m *Model) selectedNote() appdomain.Note {
	item := m.List().SelectedItem()
	if item == nil {
		return appdomain.DefaultNote()
	}
	noteItem, ok := item.(NoteItem)
	if !ok {
		return appdomain.DefaultNote()
	}
	return noteItem.Note
}

func (m *Model) selectedFolder() Folder {
	item := m.Folders.SelectedItem()
	if item == nil {
		return "misc"
	}
	return item.(Folder)
}

func (m *Model) List() *list.Model {
	folder := m.selectedFolder()
	if l, ok := m.Lists[folder]; ok {
		return l
	}
	for _, l := range m.Lists {
		return l
	}
	l := newList([]list.Item{}, m.height, m.ListStyle)
	m.Lists[folder] = l
	return l
}

func (m *Model) moveNoteDown() {
	currentPosition := m.List().Index()
	currentItem := m.List().SelectedItem()
	m.List().InsertItem(currentPosition+2, currentItem)
	m.List().RemoveItem(currentPosition)
	m.List().CursorDown()
}

func (m *Model) moveNoteUp() {
	currentPosition := m.List().Index()
	currentItem := m.List().SelectedItem()
	m.List().RemoveItem(currentPosition)
	m.List().InsertItem(currentPosition-1, currentItem)
	m.List().CursorUp()
}

func (m *Model) deleteFolder() {
	folder := m.selectedFolder()
	if folder == Folder(appdomain.DefaultNotebook) {
		return
	}

	noteList, ok := m.Lists[folder]
	if !ok {
		return
	}

	for _, item := range noteList.Items() {
		noteItem := item.(NoteItem)
		noteItem.Note.Notebook = appdomain.DefaultNotebook
		m.Lists[Folder(appdomain.DefaultNotebook)].InsertItem(0, list.Item(noteItem))
	}

	m.library.Repository.DeleteFolder(string(folder))
	delete(m.Lists, folder)

	var folderItems []list.Item
	foldersSlice := maps.Keys(m.Lists)
	slices.Sort(foldersSlice)
	for _, f := range foldersSlice {
		folderItems = append(folderItems, Folder(f))
	}
	m.Folders.SetItems(folderItems)
	for i, f := range foldersSlice {
		if f == appdomain.DefaultNotebook {
			m.Folders.Select(i)
			break
		}
	}
}

func (m *Model) createNewNoteFile() tea.Cmd {
	return func() tea.Msg {
		folder := defaultNoteFolder
		folderItem := m.Folders.SelectedItem()
		if folderItem != nil && folderItem.FilterValue() != "" {
			folder = folderItem.FilterValue()
		}

		note, err := m.library.CreateNote(folder)
		if err != nil {
			return changeStateMsg{navigatingState}
		}
		newNoteItem := NoteItem{note}

		m.List().InsertItem(m.List().Index(), newNoteItem)
		return changeStateMsg{navigatingState}
	}
}

func (m *Model) View() string {
	if m.state == quittingState {
		return ""
	}

	var (
		folder   = m.ContentStyle.Title.Render(m.selectedNote().Notebook)
		name     = m.ContentStyle.Title.Render(m.selectedNote().Title)
		titleBar = m.ListStyle.TitleBar.Render("Notes")
	)

	if m.state == contentEditingState {
		folder = m.ContentStyle.Title.Render(m.selectedNote().Notebook)
		name = m.ContentStyle.Title.Render(m.selectedNote().Title)
		if m.showPreview {
			titleBar = m.ListStyle.TitleBar.Render("Preview: " + m.selectedNote().Title)
		} else {
			titleBar = m.ListStyle.TitleBar.Render("Editing: " + m.selectedNote().Title)
		}
		contentView := m.Textarea.View()
		if m.showPreview {
			contentView = strings.ReplaceAll(m.Code.View(), "\t", strings.Repeat(" ", tabSpaces))
		}
		return lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.FoldersStyle.Base.Render(m.Folders.View()),
				m.ListStyle.Base.Render(titleBar+m.List().View()),
				lipgloss.JoinVertical(lipgloss.Top,
					lipgloss.JoinHorizontal(lipgloss.Left,
						folder,
						m.ContentStyle.Separator.Render("/"),
						name,
					),
					m.ContentStyle.Base.Render(contentView),
				),
			),
			marginStyle.Render(m.help.View(m.keys)),
		)
	}

	if m.state == editingState {
		folder = m.inputs[folderInput].View()
		name = m.inputs[nameInput].View()
	} else if m.state == creatingFolderState {
		titleBar = m.ListStyle.TitleBar.Render("New folder: " + m.createFolderInput.View())
	} else if m.state == deletingFolderState {
		titleBar = m.ListStyle.DeletedTitleBar.Render("Delete folder? (y/N)")
	} else if m.state == copyingState {
		titleBar = m.ListStyle.CopiedTitleBar.Render("Copied Note!")
	} else if m.state == deletingState {
		titleBar = m.ListStyle.DeletedTitleBar.Render("Delete Note? (y/N)")
	} else if m.List().SettingFilter() {
		titleBar = m.ListStyle.TitleBar.Render(m.List().FilterInput.View())
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.FoldersStyle.Base.Render(m.Folders.View()),
			m.ListStyle.Base.Render(titleBar+m.List().View()),
			lipgloss.JoinVertical(lipgloss.Top,
				lipgloss.JoinHorizontal(lipgloss.Left,
					folder,
					m.ContentStyle.Separator.Render("/"),
					name,
				),
				lipgloss.JoinHorizontal(lipgloss.Left,
					m.ContentStyle.LineNumber.Render(m.LineNumbers.View()),
					m.ContentStyle.Base.Render(strings.ReplaceAll(m.Code.View(), "\t", strings.Repeat(" ", tabSpaces))),
				),
			),
		),
		marginStyle.Render(m.help.View(m.keys)),
	)
}

func (m *Model) saveState() {
	s := State{
		CurrentFolder: string(m.selectedFolder()),
		CurrentNote:   m.selectedNote().File,
	}
	err := s.Save()
	if err != nil {
		panic(err.Error())
	}
}
