package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit           key.Binding
	Search         key.Binding
	ToggleHelp     key.Binding
	NewNote        key.Binding
	MoveNoteUp     key.Binding
	MoveNoteDown   key.Binding
	DeleteNote     key.Binding
	EditNote       key.Binding
	CopyNote       key.Binding
	PasteNote      key.Binding
	SetFolder      key.Binding
	RenameNote     key.Binding
	TagNote        key.Binding
	Confirm        key.Binding
	Cancel         key.Binding
	NextPane       key.Binding
	PreviousPane   key.Binding
	ChangeFolder   key.Binding
	Preview        key.Binding
	EditExternal   key.Binding
	InsertCode     key.Binding
	InsertTable    key.Binding
	InsertList     key.Binding
	InsertQuote    key.Binding
	InsertLink     key.Binding
	InsertHeader   key.Binding
	InsertRule     key.Binding
	Archive        key.Binding
	OpenNote       key.Binding
	Esc            key.Binding
	ToggleArchived key.Binding
	CreateFolder   key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit:           key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "exit")),
	Search:         key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	ToggleHelp:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	NewNote:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	MoveNoteUp:     key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "move note up")),
	MoveNoteDown:   key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "move note down")),
	DeleteNote:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	EditNote:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	CopyNote:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
	PasteNote:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "paste")),
	RenameNote:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	SetFolder:      key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "move folder")),
	TagNote:        key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tag"), key.WithDisabled()),
	Confirm:        key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
	Cancel:         key.NewBinding(key.WithKeys("N", "esc"), key.WithHelp("N", "cancel")),
	NextPane:       key.NewBinding(key.WithKeys("tab", "right"), key.WithHelp("tab", "navigate")),
	PreviousPane:   key.NewBinding(key.WithKeys("shift+tab", "left"), key.WithHelp("shift+tab", "navigate")),
	ChangeFolder:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select folder"), key.WithDisabled()),
	OpenNote:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Preview:        key.NewBinding(key.WithKeys("f2"), key.WithHelp("f2", "preview")),
	EditExternal:   key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "external edit")),
	InsertCode:     key.NewBinding(key.WithKeys("f3"), key.WithHelp("f3", "code block")),
	InsertTable:    key.NewBinding(key.WithKeys("f4"), key.WithHelp("f4", "table")),
	InsertList:     key.NewBinding(key.WithKeys("f5"), key.WithHelp("f5", "checklist")),
	InsertQuote:    key.NewBinding(key.WithKeys("f6"), key.WithHelp("f6", "quote")),
	InsertLink:     key.NewBinding(key.WithKeys("f7"), key.WithHelp("f7", "link")),
	InsertHeader:   key.NewBinding(key.WithKeys("f8"), key.WithHelp("f8", "heading")),
	InsertRule:     key.NewBinding(key.WithKeys("f9"), key.WithHelp("f9", "rule")),
	Archive:        key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
	ToggleArchived: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "show archived")),
	CreateFolder:   key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "new folder")),
	Esc:            key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextPane,
		k.Search,
		k.EditNote,
		k.DeleteNote,
		k.CopyNote,
		k.NewNote,
		k.ToggleHelp,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NewNote, k.EditNote, k.PasteNote, k.CopyNote, k.DeleteNote, k.Archive, k.ToggleArchived, k.CreateFolder},
		{k.MoveNoteDown, k.MoveNoteUp},
		{k.RenameNote, k.SetFolder, k.TagNote},
		{k.NextPane, k.PreviousPane},
		{k.Search, k.ToggleHelp, k.Quit},
	}
}
