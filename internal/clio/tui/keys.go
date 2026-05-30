package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit            key.Binding
	Search          key.Binding
	ToggleHelp      key.Binding
	TogglePreview   key.Binding
	NewNote         key.Binding
	MoveNoteUp      key.Binding
	MoveNoteDown    key.Binding
	DeleteNote      key.Binding
	EditNote        key.Binding
	CopyNote        key.Binding
	PasteNote       key.Binding
	SetFolder       key.Binding
	RenameNote      key.Binding
	TagNote         key.Binding
	SetLanguage     key.Binding
	Confirm         key.Binding
	Cancel          key.Binding
	NextPane        key.Binding
	PreviousPane    key.Binding
	ChangeFolder    key.Binding
	NewFolder       key.Binding
	DeleteFolder    key.Binding
	InsertCodeBlock key.Binding
	InsertTable     key.Binding
	InsertChecklist key.Binding
	InsertQuote     key.Binding
	InsertLink      key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit:            key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "exit")),
	Search:          key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	ToggleHelp:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	TogglePreview:   key.NewBinding(key.WithKeys("f2"), key.WithHelp("F2", "toggle preview")),
	NewNote:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	MoveNoteDown:    key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move note down")),
	MoveNoteUp:      key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move note up")),
	DeleteNote:      key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	EditNote:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	CopyNote:        key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
	PasteNote:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "paste")),
	RenameNote:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename note")),
	SetFolder:       key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "move notebook")),
	SetLanguage:     key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "set format")),
	TagNote:         key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tag"), key.WithDisabled()),
	Confirm:         key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
	Cancel:          key.NewBinding(key.WithKeys("N", "esc"), key.WithHelp("N", "cancel")),
	NextPane:        key.NewBinding(key.WithKeys("tab", "right"), key.WithHelp("tab", "navigate")),
	PreviousPane:    key.NewBinding(key.WithKeys("shift+tab", "left"), key.WithHelp("shift+tab", "navigate")),
	ChangeFolder:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "change notebook"), key.WithDisabled()),
	NewFolder:       key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "new folder")),
	DeleteFolder:    key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "delete folder")),
	InsertCodeBlock: key.NewBinding(key.WithKeys("f3"), key.WithHelp("F3", "insert code block")),
	InsertTable:     key.NewBinding(key.WithKeys("f4"), key.WithHelp("F4", "insert table")),
	InsertChecklist: key.NewBinding(key.WithKeys("f5"), key.WithHelp("F5", "insert checklist")),
	InsertQuote:     key.NewBinding(key.WithKeys("f6"), key.WithHelp("F6", "insert quote")),
	InsertLink:      key.NewBinding(key.WithKeys("f7"), key.WithHelp("F7", "insert link")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextPane,
		k.Search,
		k.EditNote,
		k.TogglePreview,
		k.NewNote,
		k.InsertCodeBlock,
		k.InsertTable,
		k.InsertChecklist,
		k.InsertQuote,
		k.InsertLink,
		k.DeleteNote,
		k.CopyNote,
		k.PasteNote,
		k.RenameNote,
		k.SetFolder,
		k.SetLanguage,
		k.NewFolder,
		k.DeleteFolder,
		k.Quit,
		k.ToggleHelp,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NewNote, k.EditNote, k.DeleteNote, k.CopyNote, k.PasteNote},
		{k.MoveNoteDown, k.MoveNoteUp},
		{k.RenameNote, k.SetFolder, k.SetLanguage, k.NewFolder, k.DeleteFolder},
		{k.NextPane, k.PreviousPane, k.TogglePreview},
		{k.Search, k.Quit, k.ToggleHelp},
		{k.InsertCodeBlock, k.InsertTable, k.InsertChecklist, k.InsertQuote, k.InsertLink},
	}
}
