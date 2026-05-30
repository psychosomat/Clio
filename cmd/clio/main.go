package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	appconfig "clio/internal/clio/config"
	appdomain "clio/internal/clio/domain"
	appservice "clio/internal/clio/service"
	"clio/internal/clio/tui"

	"github.com/mattn/go-isatty"
)

var helpText = strings.TrimSpace(`
Clio is a notes application for your terminal.
https://github.com/maaslalani/clio

Usage:
  clio           - interactive mode
  clio list      - list all notes
  clio <note>    - print note to stdout

Create:
  clio < note.md                 - save note from stdin
  clio Work/meeting.md < note.md - save note with name`)

func main() {
	runCLI(os.Args[1:])
}

func runCLI(args []string) {
	config := appconfig.Load()
	library := appservice.New(config)
	notes, err := library.Bootstrap()
	if err != nil {
		fmt.Println("Unable to load notes", err)
		return
	}

	stdin := readStdin()
	if stdin != "" {
		if _, err := library.SaveNote(stdin, args, notes); err != nil {
			fmt.Println("Unable to save note", err)
		}
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "list":
			listNotes(notes)
		case "-h", "--help":
			fmt.Println(helpText)
		default:
			note := library.FindNote(args[0], notes)
			fmt.Print(library.RenderContent(note, isatty.IsTerminal(os.Stdout.Fd())))
		}
		return
	}

	err = tui.Run(config, library, notes)
	if err != nil {
		fmt.Println("Alas, there's been an error", err)
	}
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
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			return ""
		}
	}

	return b.String()
}

func listNotes(notes []appdomain.Note) {
	for _, note := range notes {
		fmt.Println(note.File)
	}
}
