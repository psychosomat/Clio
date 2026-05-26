package app

import (
	"context"
	"time"

	"clio/internal/notes"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadNotesCmd(store notes.Store) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		notesList, err := store.List(ctx, notes.ListOptions{IncludeArchived: true})
		if err != nil {
			return MsgError{Err: err}
		}

		return MsgNotesLoaded{Notes: notesList}
	}
}

func SaveNoteCmd(store notes.Store, note notes.Note, seq int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		saved, err := store.Save(ctx, note)
		if err != nil {
			return MsgError{Err: err}
		}

		return MsgNoteSaved{Note: saved, Seq: seq}
	}
}

func DeleteNoteCmd(store notes.Store, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := store.MoveToTrash(ctx, id)
		if err != nil {
			return MsgError{Err: err}
		}

		return MsgLoadNotes{}
	}
}

func SetStatusTimeoutCmd(id int, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(delay)
		return MsgStatusClear{ID: id}
	}
}

func AutosaveCmd(seq int, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(delay)
		return MsgAutosaveTick{Seq: seq}
	}
}

func OpenLinkCmd(opener LinkOpener, target string) tea.Cmd {
	return func() tea.Msg {
		if err := opener.Open(target); err != nil {
			return MsgError{Err: err}
		}
		return MsgLinkOpened{Target: target}
	}
}
