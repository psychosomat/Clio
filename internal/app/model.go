package app

import (
	"clio/internal/markdownpreview"
	"clio/internal/notes"
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

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

type pane int

const (
	notePane pane = iota
	contentPane
	folderPane
)

const maxPane = 3

type appMode int

const (
	browsingMode appMode = iota
	editorMode
	searchMode
	helpMode
)

type browseState int

const (
	navigating browseState = iota
	copyVisual
	deleteConfirm
)

const (
	defaultFolder  = "notes"
	defaultTitle   = "Untitled"
	listWidth      = 33
	foldersWidth   = 20
	lineNumWidth   = 5
	previewPadding = 15
)

var defaultNote = Note{
	Title:  defaultTitle,
	Folder: defaultFolder,
	Date:   time.Now(),
	Tags:   make([]string, 0),
}

type Note struct {
	ID       string
	Tags     []string
	Folder   string
	Date     time.Time
	Favorite bool
	Title    string
	File     string
	Body     string
	FilePath string
}

func NoteFromNotesNote(n notes.Note) Note {
	folder := n.Folder
	if folder == "" {
		folder = defaultFolder
	}
	return Note{
		ID:       n.ID,
		Tags:     n.Tags,
		Folder:   folder,
		Date:     n.UpdatedAt,
		Title:    n.DisplayTitle(),
		Body:     n.Body,
		File:     n.ID,
		FilePath: n.Path,
	}
}

func (n Note) FilterValue() string {
	return n.Folder + "/" + n.Title + "\n" + "#" + strings.Join(n.Tags, " #")
}

func (n Note) String() string {
	return fmt.Sprintf("%s/%s", n.Folder, n.Title)
}

func (n Note) NotePath() string {
	if n.FilePath != "" {
		return n.FilePath
	}
	return filepath.Join(n.Folder, n.File)
}

type Folder string

func (f Folder) FilterValue() string { return string(f) }

type Model struct {
	config Config
	keys   KeyMap
	help   help.Model
	height int
	width  int
	Store  notes.Store

	Lists   map[Folder]*list.Model
	Folders list.Model

	Notes    []notes.Note
	Filtered []notes.Note

	Code        viewport.Model
	LineNumbers viewport.Model

	pane        pane
	Mode        appMode
	browseState browseState

	ListStyle    NotesBaseStyle
	FoldersStyle FoldersBaseStyle
	ContentStyle ContentBaseStyle

	CurrentNote   notes.Note
	EditorBody    textarea.Model
	EditorPreview bool
	EditorDirty   bool
	Saving        bool
	AutosaveSeq   int
	SavingSeq     int

	StatusMsg   string
	StatusMsgID int
	StatusIsErr bool

	SearchInput  textinput.Model
	ShowArchived bool

	RenameInput  textinput.Model
	RenameActive bool
	FolderInput  textinput.Model
	FolderActive bool

	PreviewViewport   viewport.Model
	PreviewRenderer   *markdownpreview.Renderer
	PreviewResult     markdownpreview.RenderResult
	PreviewLinks      []markdownpreview.LinkTarget
	PreviewLinkIndex  int
	PreviewRenderKey  string
	LastRenderedBody  string
	LastRenderedWidth int
	LastRenderedANSI  string
	Opener            LinkOpener
}

func NewModel(store notes.Store, config Config, notesList []notes.Note) *Model {
	state := ReadState()

	folders := make(map[Folder][]list.Item)
	for _, n := range notesList {
		note := NoteFromNotesNote(n)
		folders[Folder(note.Folder)] = append(folders[Folder(note.Folder)], list.Item(note))
	}

	defaultStyles := DefaultStyles(config)

	var folderItems []list.Item
	foldersSlice := maps.Keys(folders)
	slices.Sort(foldersSlice)
	for _, folder := range foldersSlice {
		folderItems = append(folderItems, list.Item(folder))
	}
	if len(folderItems) <= 0 {
		folderItems = append(folderItems, list.Item(Folder(defaultFolder)))
	}
	folderList := list.New(folderItems, folderDelegate{defaultStyles.Folders.Blurred}, 0, 0)
	folderList.Title = "Folders"
	folderList.SetShowHelp(false)
	folderList.SetFilteringEnabled(false)
	folderList.SetShowStatusBar(false)
	folderList.DisableQuitKeybindings()
	folderList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color(config.GrayColor))
	folderList.SetStatusBarItemName("folder", "folders")

	for idx, folder := range foldersSlice {
		if string(folder) == state.CurrentFolder {
			folderList.Select(idx)
			break
		}
	}

	content := viewport.New(80, 0)
	ln := viewport.New(0, 0)

	lists := make(map[Folder]*list.Model)
	currentFolder := folderList.SelectedItem().(Folder)
	for folder, items := range folders {
		noteList := newList(items, 20, defaultStyles.Notes.Focused, config)
		if folder == currentFolder {
			for idx, item := range noteList.Items() {
				if n, ok := item.(Note); ok && n.File == state.CurrentNote {
					noteList.Select(idx)
					break
				}
			}
		}
		lists[folder] = noteList
	}

	si := textinput.New()
	si.Placeholder = "Type to filter..."
	si.Prompt = " / "
	si.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor)).Bold(true)
	si.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor))
	si.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.WhiteColor))
	si.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.GrayColor))

	eb := textarea.New()
	eb.Placeholder = "Start writing (Markdown)..."
	eb.ShowLineNumbers = true

	vp := viewport.New(0, 0)

	ri := textinput.New()
	ri.Prompt = "Title: "
	ri.Placeholder = "Note title"
	ri.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor)).Bold(true)
	ri.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor))
	ri.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.WhiteColor))
	ri.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.GrayColor))

	fi := textinput.New()
	fi.Prompt = "Folder: "
	fi.Placeholder = "Folder name"
	fi.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor)).Bold(true)
	fi.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(config.PrimaryColor))
	fi.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.WhiteColor))
	fi.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(config.GrayColor))

	m := &Model{
		Lists:           lists,
		Folders:         folderList,
		Code:            content,
		LineNumbers:     ln,
		ContentStyle:    defaultStyles.Content.Blurred,
		ListStyle:       defaultStyles.Notes.Focused,
		FoldersStyle:    defaultStyles.Folders.Blurred,
		keys:            DefaultKeyMap,
		help:            help.New(),
		config:          config,
		Store:           store,
		Notes:           notesList,
		Filtered:        notesList,
		RenameInput:     ri,
		FolderInput:     fi,
		PreviewRenderer: markdownpreview.NewRenderer(),
		Opener:          NewSystemLinkOpener(),
		SearchInput:     si,
		EditorBody:      eb,
		PreviewViewport: vp,
		Mode:            browsingMode,
		pane:            notePane,
		browseState:     navigating,
	}

	return m
}

type updateContentMsg Note

func (m *Model) updateContent() tea.Cmd {
	return func() tea.Msg {
		return updateContentMsg(m.selectedNote())
	}
}

type updateFoldersMsg struct {
	items               []list.Item
	selectedFolderIndex int
}

func (m *Model) updateFolders() tea.Cmd {
	return func() tea.Msg {
		return m.updateFoldersView()
	}
}

type changeStateMsg struct{ newState browseState }

func changeState(newState browseState) tea.Cmd {
	return func() tea.Msg {
		return changeStateMsg{newState}
	}
}

func (m *Model) Init() tea.Cmd {
	rand.Seed(time.Now().Unix())
	m.Folders.Styles.Title = m.FoldersStyle.Title
	m.Folders.Styles.TitleBar = m.FoldersStyle.TitleBar
	m.updateKeyMap()
	return LoadNotesCmd(m.Store)
}

func (m *Model) InitNewNote() {
	m.CurrentNote = notes.Note{}
	m.EditorBody.SetValue("")
	m.EditorBody.Focus()
	m.EditorPreview = false
	m.EditorDirty = false
	m.Saving = false
	m.AutosaveSeq = 0
	m.SavingSeq = 0
	m.Mode = editorMode
}

func (m *Model) selectedNote() Note {
	if m.List() == nil {
		return defaultNote
	}
	item := m.List().SelectedItem()
	if item == nil {
		return defaultNote
	}
	return item.(Note)
}

func (m *Model) selectedFolder() Folder {
	item := m.Folders.SelectedItem()
	if item == nil {
		return Folder(defaultFolder)
	}
	return item.(Folder)
}

func (m *Model) List() *list.Model {
	f := m.selectedFolder()
	if _, ok := m.Lists[f]; !ok {
		m.Lists[f] = newList([]list.Item{}, m.height, m.ListStyle, m.config)
	}
	return m.Lists[f]
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

func (m *Model) moveNoteDown() tea.Cmd {
	l := m.List()
	if l == nil {
		return nil
	}
	currentPosition := l.Index()
	if currentPosition < 0 || currentPosition >= len(l.Items())-1 {
		return nil
	}
	currentItem := l.SelectedItem()
	if currentItem == nil {
		return nil
	}
	l.RemoveItem(currentPosition)
	l.InsertItem(currentPosition+1, currentItem)
	l.CursorDown()
	return m.saveListOrder()
}

func (m *Model) moveNoteUp() tea.Cmd {
	l := m.List()
	if l == nil {
		return nil
	}
	currentPosition := l.Index()
	if currentPosition <= 0 {
		return nil
	}
	currentItem := l.SelectedItem()
	if currentItem == nil {
		return nil
	}
	l.RemoveItem(currentPosition)
	l.InsertItem(currentPosition-1, currentItem)
	l.CursorUp()
	return m.saveListOrder()
}

func (m *Model) saveListOrder() tea.Cmd {
	l := m.List()
	if l == nil {
		return nil
	}
	// Build a fast ID-to-note index from m.Notes
	noteIdx := make(map[string]*notes.Note)
	for i := range m.Notes {
		noteIdx[m.Notes[i].ID] = &m.Notes[i]
	}

	// Assign positions to all notes in the current folder based on list order
	var toSave []notes.Note
	for i, item := range l.Items() {
		appNote := item.(Note)
		if n, ok := noteIdx[appNote.ID]; ok && n.Position != i {
			n.Position = i
			toSave = append(toSave, *n)
		}
	}
	if len(toSave) == 0 {
		return nil
	}

	// Also sync positions into m.Filtered so the UI stays consistent
	for i := range m.Filtered {
		if n, ok := noteIdx[m.Filtered[i].ID]; ok {
			m.Filtered[i].Position = n.Position
		}
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, n := range toSave {
			if _, err := m.Store.Save(ctx, n); err != nil {
				return MsgError{Err: fmt.Errorf("failed to save position: %w", err)}
			}
		}
		return MsgLoadNotes{}
	}
}

func (m *Model) createFolder() tea.Cmd {
	return func() tea.Msg {
		m.FolderInput.SetValue("")
		m.FolderActive = true
		m.FolderInput.Focus()
		return nil
	}
}

func (m *Model) ensureFolderExists(f Folder) {
	if _, ok := m.Lists[f]; !ok {
		m.Lists[f] = newList([]list.Item{}, m.height, m.ListStyle, m.config)
	}
	found := false
	for _, item := range m.Folders.Items() {
		if existing, ok := item.(Folder); ok && existing == f {
			found = true
			break
		}
	}
	if !found {
		m.Folders.InsertItem(0, list.Item(f))
	}
}

func (m *Model) createNewNote() tea.Cmd {
	return func() tea.Msg {
		folder := defaultFolder
		folderItem := m.Folders.SelectedItem()
		if folderItem != nil && folderItem.FilterValue() != "" {
			folder = folderItem.FilterValue()
		}

		now := time.Now()
		note := notes.Note{
			Title:     defaultTitle,
			Body:      "",
			Folder:    folder,
			CreatedAt: now,
			UpdatedAt: now,
		}
		note.ID = now.Format("20060102150405") + "-" + fmt.Sprintf("%d", rand.Intn(1000000))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		saved, err := m.Store.Save(ctx, note)
		if err != nil {
			return MsgError{Err: fmt.Errorf("failed to create note: %w", err)}
		}

		appNote := NoteFromNotesNote(saved)
		if m.List() != nil {
			m.List().InsertItem(0, appNote)
		}
		return MsgNoteCreated{Note: saved}
	}
}

func (m *Model) openNote(note notes.Note) {
	m.CurrentNote = note
	m.EditorBody.SetValue(note.Body)
	m.EditorBody.Focus()
	m.EditorDirty = false
	m.EditorPreview = false
	m.Mode = editorMode
}

func (m *Model) updateKeyMap() {
	if m.List() == nil {
		return
	}
	hasItems := len(m.List().VisibleItems()) > 0
	isFiltering := m.List().FilterState() == list.Filtering
	m.keys.DeleteNote.SetEnabled(hasItems && !isFiltering)
	m.keys.CopyNote.SetEnabled(hasItems && !isFiltering)
	m.keys.EditNote.SetEnabled(hasItems && !isFiltering)
	m.keys.NewNote.SetEnabled(!isFiltering)
	m.keys.ChangeFolder.SetEnabled(m.pane == folderPane)
}

func (m *Model) saveState() {
	s := State{
		CurrentFolder: string(m.selectedFolder()),
		CurrentNote:   m.selectedNote().File,
	}
	_ = s.Save()
}

// -- Content / Preview helpers --

func (m *Model) writeLineNumbers(n int) {
	if n <= 0 {
		m.LineNumbers.SetContent("")
		return
	}
	var ln strings.Builder
	for i := 1; i <= n; i++ {
		if i > 1 {
			ln.WriteByte('\n')
		}
		fmt.Fprintf(&ln, "%3d ", i)
	}
	m.LineNumbers.SetContent(ln.String())
}

func (m *Model) updateMarkdownPreview() {
	body, width, ok := m.previewBodyAndWidth()
	if !ok {
		m.PreviewLinks = nil
		m.PreviewLinkIndex = -1
		m.LastRenderedBody = ""
		m.LastRenderedWidth = 0
		m.LastRenderedANSI = ""
		m.PreviewViewport.SetContent("")
		m.LineNumbers.SetContent("")
		return
	}
	if width < 20 {
		width = 20
	}

	prevBody := m.LastRenderedBody
	cacheKey := fmt.Sprintf("%d:%d:%s", width, m.PreviewLinkIndex, body)
	if m.LastRenderedBody == body && m.LastRenderedWidth == width && m.PreviewRenderKey == cacheKey {
		return
	}

	if strings.TrimSpace(body) == "" {
		m.LastRenderedBody = body
		m.LastRenderedWidth = width
		m.PreviewRenderKey = ""
		m.PreviewLinks = nil
		m.PreviewLinkIndex = -1
		m.LastRenderedANSI = ""
		m.PreviewViewport.SetContent("")
		m.LineNumbers.SetContent("")
		return
	}

	if len(m.PreviewLinks) == 0 {
		m.PreviewLinkIndex = -1
	} else if m.PreviewLinkIndex >= len(m.PreviewLinks) {
		m.PreviewLinkIndex = 0
	}

	if body != prevBody {
		m.PreviewViewport.GotoTop()
	}

	result, err := m.PreviewRenderer.Render(body, markdownpreview.RenderOptions{
		Width:           width,
		TerminalLinks:   true,
		ActiveLinkIndex: m.PreviewLinkIndex,
	})
	if err != nil {
		m.LastRenderedANSI = "preview error"
		m.PreviewViewport.SetContent(m.LastRenderedANSI)
		return
	}

	m.LastRenderedBody = body
	m.LastRenderedWidth = width
	m.PreviewRenderKey = cacheKey
	m.PreviewResult = result
	m.PreviewLinks = result.Links
	if len(m.PreviewLinks) == 0 {
		m.PreviewLinkIndex = -1
	} else if m.PreviewLinkIndex < 0 {
		m.PreviewLinkIndex = 0
	}
	m.LastRenderedANSI = result.ANSI
	m.PreviewViewport.SetContent(result.ANSI)
	m.writeLineNumbers(lipgloss.Height(result.ANSI))
	m.LineNumbers.SetYOffset(m.PreviewViewport.YOffset)
}

func (m *Model) previewBodyAndWidth() (string, int, bool) {
	switch m.Mode {
	case editorMode:
		width := m.PreviewViewport.Width
		if width < 20 {
			width = m.width - 50
			if width < 20 {
				width = 20
			}
		}
		return m.EditorBody.Value(), width, true
	case browsingMode, searchMode:
		if len(m.Filtered) == 0 || m.List().Index() < 0 {
			return "", 0, false
		}
		width := m.width - m.List().Width() - m.Folders.Width() - 25
		if width < 20 {
			width = 20
		}
		return m.selectedNote().Body, width, true
	default:
		return "", 0, false
	}
}

func (m *Model) resetPreviewSelection() {
	m.PreviewLinkIndex = -1
	m.PreviewRenderKey = ""
	m.LastRenderedBody = ""
	m.PreviewViewport.GotoTop()
	m.PreviewViewport.SetContent("")
	m.LineNumbers.SetContent("")
}

func (m *Model) cyclePreviewLink(delta int) {
	if len(m.PreviewLinks) == 0 {
		m.PreviewLinkIndex = -1
		return
	}
	if m.PreviewLinkIndex < 0 {
		if delta >= 0 {
			m.PreviewLinkIndex = 0
		} else {
			m.PreviewLinkIndex = len(m.PreviewLinks) - 1
		}
		return
	}
	m.PreviewLinkIndex = (m.PreviewLinkIndex + delta + len(m.PreviewLinks)) % len(m.PreviewLinks)
}

func (m *Model) openPreviewLink() tea.Cmd {
	if m.PreviewLinkIndex < 0 || m.PreviewLinkIndex >= len(m.PreviewLinks) {
		return nil
	}
	link := m.PreviewLinks[m.PreviewLinkIndex]
	switch link.Kind {
	case markdownpreview.LinkKindAnchor:
		return nil
	case markdownpreview.LinkKindWiki:
		return nil
	default:
		target := link.URL
		return OpenLinkCmd(m.Opener, target)
	}
}

// -- Editor helpers --

func (m *Model) editorNoteForSave() (notes.Note, bool) {
	note := m.CurrentNote
	note.Body = m.EditorBody.Value()
	if note.ID == "" && strings.TrimSpace(note.Body) == "" {
		return note, false
	}
	return note, true
}

func (m *Model) restoreEditorCursor(line, col int) {
	lineCount := m.EditorBody.LineCount()
	if lineCount == 0 {
		return
	}
	if line < 0 {
		line = 0
	}
	if line >= lineCount {
		line = lineCount - 1
	}
	m.EditorBody.CursorStart()
	for m.EditorBody.Line() > line {
		m.EditorBody.CursorUp()
	}
	m.EditorBody.SetCursor(col)
}

// -- Folder update helpers --

func (m *Model) rebuildFolderLists() {
	folderItems := make(map[Folder][]list.Item)
	for _, notesItem := range m.Notes {
		if notesItem.Archived && !m.ShowArchived {
			continue
		}
		f := Folder(notesItem.Folder)
		if f == "" {
			f = Folder(defaultFolder)
		}
		folderItems[f] = append(folderItems[f], NoteFromNotesNote(notesItem))
	}

	for f, items := range folderItems {
		if _, ok := m.Lists[f]; !ok {
			m.Lists[f] = newList([]list.Item{}, m.height, m.ListStyle, m.config)
		}
		m.Lists[f].SetItems(items)
	}

	// Remove lists for folders that no longer have notes
	for f := range m.Lists {
		if _, ok := folderItems[f]; !ok {
			delete(m.Lists, f)
		}
	}

	var folderNames []list.Item
	foldersSlice := maps.Keys(folderItems)
	slices.Sort(foldersSlice)
	for _, folder := range foldersSlice {
		folderNames = append(folderNames, Folder(folder))
	}
	setItemsCmd := m.Folders.SetItems(folderNames)

	selectedIndex := 0
	for i, f := range foldersSlice {
		if f == m.selectedFolder() {
			selectedIndex = i
			break
		}
	}
	m.Folders.Select(selectedIndex)
	m.Folders, _ = m.Folders.Update(nil)
	_ = setItemsCmd
}

func (m *Model) updateFoldersView() tea.Msg {
	var selectedFolder Folder
	selectedFolderIndex := m.Folders.Index()

	for folder, li := range m.Lists {
		for i := 0; i < len(li.Items()); i++ {
			item := li.Items()[i]
			n, ok := item.(Note)
			if !ok {
				continue
			}
			f := Folder(n.Folder)
			if _, ok := m.Lists[f]; !ok {
				m.Lists[f] = newList([]list.Item{}, m.height, m.ListStyle, m.config)
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

// -- Styling / Content display --

func (m *Model) displayKeyHint(hints []keyHint) {
	var ln strings.Builder
	for range hints {
		ln.WriteString("  ~ \n")
	}
	m.LineNumbers.SetContent(ln.String())
	var s strings.Builder
	for _, h := range hints {
		s.WriteString(fmt.Sprintf("%s %s\n",
			m.ContentStyle.EmptyHintKey.Render(h.binding.Help().Key),
			m.ContentStyle.EmptyHint.Render("• "+h.help),
		))
	}
	m.Code.SetContent(s.String())
}

func (m *Model) displayError(err string) {
	m.LineNumbers.SetContent(" ~ ")
	m.Code.SetContent(m.ContentStyle.EmptyHint.Render(err))
}

func (m *Model) renderPreviewInPane(body string) {
	previewWidth := m.width - listWidth - foldersWidth - previewPadding
	if previewWidth < 20 {
		previewWidth = 20
	}
	if strings.TrimSpace(body) == "" {
		m.displayKeyHint(m.noContentHints())
		return
	}
	result, err := m.PreviewRenderer.Render(body, markdownpreview.RenderOptions{
		Width:         previewWidth,
		TerminalLinks: true,
	})
	if err != nil {
		m.displayError("Preview error")
		return
	}
	m.writeLineNumbers(lipgloss.Height(result.ANSI))
	m.Code.SetContent(result.ANSI)
}

type keyHint struct {
	binding key.Binding
	help    string
}

func (m *Model) noContentHints() []keyHint {
	return []keyHint{
		{m.keys.EditNote, "edit contents"},
		{m.keys.RenameNote, "rename"},
	}
}

// -- Pane style update --

func (m *Model) updatePaneStyles(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	defStyles := DefaultStyles(m.config)

	switch m.pane {
	case folderPane:
		m.ListStyle = defStyles.Notes.Blurred
		m.ContentStyle = defStyles.Content.Blurred
		m.FoldersStyle = defStyles.Folders.Focused
		m.Folders, cmd = m.Folders.Update(msg)
		cmds = append(cmds, cmd)
	case notePane:
		m.ListStyle = defStyles.Notes.Focused
		m.ContentStyle = defStyles.Content.Blurred
		m.FoldersStyle = defStyles.Folders.Blurred
		if m.List() != nil {
			*m.List(), cmd = m.List().Update(msg)
			cmds = append(cmds, cmd)
		}
	case contentPane:
		m.ListStyle = defStyles.Notes.Blurred
		m.ContentStyle = defStyles.Content.Focused
		m.FoldersStyle = defStyles.Folders.Blurred
		m.Code, cmd = m.Code.Update(msg)
		cmds = append(cmds, cmd)
		m.LineNumbers, cmd = m.LineNumbers.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.List() != nil {
		m.List().SetDelegate(noteDelegate{m.ListStyle, m.browseState})
	}
	m.Folders.SetDelegate(folderDelegate{m.FoldersStyle})
	m.Folders.Styles.TitleBar = m.FoldersStyle.TitleBar
	m.Folders.Styles.Title = m.FoldersStyle.Title

	return tea.Batch(cmds...)
}

func newList(items []list.Item, height int, styles NotesBaseStyle, cfg Config) *list.Model {
	l := list.New(items, noteDelegate{styles, navigating}, 25, height)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowTitle(false)
	l.Styles.StatusBar = lipgloss.NewStyle().Margin(1, 2).Foreground(lipgloss.Color(cfg.GrayColor)).MaxWidth(listWidth - 2)
	l.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color(cfg.BlackColor)).MaxWidth(listWidth - 2)
	l.FilterInput.Prompt = "Find: "
	l.FilterInput.PromptStyle = styles.Title
	l.SetStatusBarItemName("note", "notes")
	l.DisableQuitKeybindings()
	l.Styles.Title = styles.Title
	l.Styles.TitleBar = styles.TitleBar
	return &l
}
