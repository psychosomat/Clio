package app

import "clio/internal/notes"

type MsgLoadNotes struct{}

type MsgNotesLoaded struct {
	Notes []notes.Note
}

type MsgNoteSaved struct {
	Note notes.Note
	Seq  int
}

type MsgAutosaveTick struct {
	Seq int
}

type MsgStatusClear struct {
	ID int
}

type MsgError struct {
	Err error
}

type MsgNoteCreated struct {
	Note notes.Note
}

type MsgLinkOpened struct {
	Target string
}
