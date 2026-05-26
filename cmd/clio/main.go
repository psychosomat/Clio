package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clio/internal/app"
	"clio/internal/notes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

var version = "dev"

var helpText = strings.TrimSpace(`
Clio is a notes app for your terminal.

Usage:
  clio           - interactive mode
  clio list      - list all notes
  clio <search>  - find and print a note

Create:
  clio new       - create a new note
`)

func main() {
	runCLI(os.Args[1:])
}

func runCLI(args []string) {
	config := app.ReadConfig()
	store := initStore(config)

	stdin := readStdin()
	if stdin != "" {
		saveFromStdin(stdin, args, store)
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "list":
			listNotes(store)
		case "new":
			runInteractive(store, config, true)
		case "-h", "--help":
			fmt.Println(helpText)
		default:
			findAndPrint(store, args[0])
		}
		return
	}

	runInteractive(store, config, false)
}

func initStore(config app.Config) *notes.FileStore {
	notesDir := filepath.Join(config.Home, "notes")
	trashDir := filepath.Join(config.Home, "trash")
	store, err := notes.NewFileStore(notesDir, trashDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing store: %v\n", err)
		os.Exit(1)
	}
	return store
}

func readStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return ""
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder
	for {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}

func saveFromStdin(content string, args []string, store *notes.FileStore) {
	title := "stdin"
	if len(args) > 0 {
		title = strings.Join(args, " ")
	}

	note := notes.Note{
		Title: title,
		Body:  content,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	saved, err := store.Save(ctx, note)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving note: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Saved:", saved.ID)
}

func listNotes(store *notes.FileStore) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allNotes, err := store.List(ctx, notes.ListOptions{IncludeArchived: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing notes: %v\n", err)
		os.Exit(1)
		return
	}

	for _, note := range allNotes {
		fmt.Println(note.DisplayTitle())
	}
}

type noteItem struct {
	title string
	note  notes.Note
}

type noteSource struct {
	items []noteItem
}

func (s noteSource) String(i int) string { return s.items[i].title }
func (s noteSource) Len() int            { return len(s.items) }

func findAndPrint(store *notes.FileStore, search string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allNotes, err := store.List(ctx, notes.ListOptions{IncludeArchived: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
		return
	}

	var items []noteItem
	for _, n := range allNotes {
		items = append(items, noteItem{title: n.DisplayTitle(), note: n})
	}

	src := noteSource{items: items}
	matches := fuzzy.FindFrom(search, src)
	if len(matches) > 0 {
		n := items[matches[0].Index].note
		fmt.Print(n.Body)
		return
	}
	os.Exit(1)
}

func runInteractive(store *notes.FileStore, config app.Config, directEdit bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allNotes, err := store.List(ctx, notes.ListOptions{IncludeArchived: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading notes: %v\n", err)
		os.Exit(1)
		return
	}

	model := app.NewModel(store, config, allNotes)
	if directEdit {
		model.InitNewNote()
	}

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
