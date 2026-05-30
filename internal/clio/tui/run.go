package tui

import (
	appconfig "clio/internal/clio/config"
	appdomain "clio/internal/clio/domain"
	appservice "clio/internal/clio/service"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func Run(config appconfig.App, library *appservice.Service, notes []appdomain.Note) error {
	if len(notes) == 0 {
		notes = append(notes, appdomain.DefaultNote())
	}
	state := readState()

	folders := make(map[Folder][]list.Item)
	for _, note := range notes {
		folders[Folder(note.Notebook)] = append(folders[Folder(note.Notebook)], list.Item(NoteItem{note}))
	}

	defaultStyles := DefaultStyles(config)

	var folderItems []list.Item
	foldersSlice := maps.Keys(folders)
	slices.Sort(foldersSlice)
	for _, folder := range foldersSlice {
		folderItems = append(folderItems, list.Item(folder))
	}
	if len(folderItems) <= 0 {
		folderItems = append(folderItems, list.Item(Folder(defaultNoteFolder)))
	}
	folderList := list.New(folderItems, folderDelegate{defaultStyles.Folders.Blurred}, 0, 0)
	folderList.Title = "Notebooks"

	folderList.SetShowHelp(false)
	folderList.SetFilteringEnabled(false)
	folderList.SetShowStatusBar(false)
	folderList.DisableQuitKeybindings()
	folderList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color(config.GrayColor))
	folderList.SetStatusBarItemName("notebook", "notebooks")

	for idx, folder := range foldersSlice {
		if string(folder) == state.CurrentFolder {
			folderList.Select(idx)
			break
		}
	}

	content := viewport.New(80, 0)

	lists := map[Folder]*list.Model{}

	currentFolder := folderList.SelectedItem().(Folder)
	for folder, items := range folders {
		noteList := newList(items, 20, defaultStyles.Notes.Focused)
		if folder == currentFolder {
			for idx, item := range noteList.Items() {
				if s, ok := item.(NoteItem); ok && s.Note.File == state.CurrentNote {
					noteList.Select(idx)
					break
				}
			}
		}
		lists[folder] = noteList
	}

	m := &Model{
		Lists:        lists,
		Folders:      folderList,
		Code:         content,
		Textarea:     newTextarea(),
		showPreview:  true,
		ContentStyle: defaultStyles.Content.Blurred,
		ListStyle:    defaultStyles.Notes.Focused,
		FoldersStyle: defaultStyles.Folders.Blurred,
		keys:         DefaultKeyMap,
		help:         help.New(),
		config:       config,
		inputs: []textinput.Model{
			newTextInput(defaultNoteFolder + " "),
			newTextInput(defaultNoteName + " "),
			newTextInput(""),
		},
		tagsInput:        newTextInput("Tags"),
		createFolderInput: newTextInput("Folder name"),
		library:          library,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return err
	}
	fm, ok := model.(*Model)
	if !ok {
		return err
	}
	var allNotes []list.Item
	for _, list := range fm.Lists {
		allNotes = append(allNotes, list.Items()...)
	}
	allNoteItems := make([]appdomain.Note, 0, len(allNotes))
	for _, item := range allNotes {
		noteItem, ok := item.(NoteItem)
		if ok {
			allNoteItems = append(allNoteItems, noteItem.Note)
		}
	}
	return library.Persist(allNoteItems)
}

func newList(items []list.Item, height int, styles NotesBaseStyle) *list.Model {
	noteList := list.New(items, noteDelegate{styles, navigatingState}, 25, height)
	noteList.SetShowHelp(false)
	noteList.SetShowFilter(false)
	noteList.SetShowTitle(false)
	noteList.Styles.StatusBar = lipgloss.NewStyle().Margin(1, 2).Foreground(lipgloss.Color("240")).MaxWidth(35 - 2)
	noteList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color("8")).MaxWidth(35 - 2)
	noteList.FilterInput.Prompt = "Find: "
	noteList.FilterInput.PromptStyle = styles.Title
	noteList.SetStatusBarItemName("note", "notes")
	noteList.DisableQuitKeybindings()
	noteList.Styles.Title = styles.Title
	noteList.Styles.TitleBar = styles.TitleBar

	return &noteList
}

func newTextInput(placeholder string) textinput.Model {
	i := textinput.New()
	i.Prompt = ""
	i.PromptStyle = lipgloss.NewStyle().Margin(0, 1)
	i.Placeholder = placeholder
	return i
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Placeholder = "Start writing..."
	t.ShowLineNumbers = true
	t.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("237"))
	t.BlurredStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("237"))
	t.FocusedStyle.Base = lipgloss.NewStyle().Margin(0, 1)
	t.BlurredStyle.Base = lipgloss.NewStyle().Margin(0, 1)
	t.CharLimit = 0
	t.MaxHeight = 0
	return t
}
