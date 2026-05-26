package app

import "strings"

const snippetCursor = "<++>"

type editorSnippet struct {
	status string
	body   string
}

func snippetForKey(msg string) (editorSnippet, bool) {
	switch msg {
	case "f3":
		return editorSnippet{
			status: "Inserted code block",
			body:   "\n```text\n" + snippetCursor + "\n```\n",
		}, true
	case "f4":
		return editorSnippet{
			status: "Inserted table",
			body:   "\n| Column | Value |\n| --- | --- |\n| " + snippetCursor + " |  |\n",
		}, true
	case "f5":
		return editorSnippet{
			status: "Inserted checklist",
			body:   "\n- [ ] " + snippetCursor + "\n- [ ] \n",
		}, true
	case "f6":
		return editorSnippet{
			status: "Inserted quote",
			body:   "\n> " + snippetCursor + "\n",
		}, true
	case "f7":
		return editorSnippet{
			status: "Inserted link",
			body:   "[" + snippetCursor + "](https://)",
		}, true
	case "f8":
		return editorSnippet{
			status: "Inserted heading",
			body:   "\n## " + snippetCursor + "\n",
		}, true
	case "f9":
		return editorSnippet{
			status: "Inserted horizontal rule",
			body:   "\n---\n",
		}, true
	default:
		return editorSnippet{}, false
	}
}

func insertSnippet(value string, line, column int, snippet string) (string, int, int) {
	cursorIndex := strings.Index(snippet, snippetCursor)
	if cursorIndex == -1 {
		cursorIndex = len(snippet)
	}

	cleanSnippet := strings.Replace(snippet, snippetCursor, "", 1)
	insertAt := editorCursorIndex(value, line, column)

	runes := []rune(value)
	updated := string(runes[:insertAt]) + cleanSnippet + string(runes[insertAt:])
	cursorAbs := insertAt + len([]rune(snippet[:cursorIndex]))

	cursorLine, cursorCol := cursorLineColumn(string([]rune(updated)), cursorAbs)

	return updated, cursorLine, cursorCol
}

func editorCursorIndex(value string, line, column int) int {
	lines := strings.Split(value, "\n")
	if len(lines) == 0 {
		return 0
	}

	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = len(lines) - 1
	}

	index := 0
	for i := 0; i < line; i++ {
		index += len([]rune(lines[i])) + 1
	}

	lineRunes := []rune(lines[line])
	if column < 0 {
		column = 0
	}
	if column > len(lineRunes) {
		column = len(lineRunes)
	}

	return index + column
}

func cursorLineColumn(value string, abs int) (int, int) {
	if abs < 0 {
		abs = 0
	}

	runes := []rune(value)
	if abs > len(runes) {
		abs = len(runes)
	}

	line := 0
	col := 0
	for i := 0; i < abs; i++ {
		if runes[i] == '\n' {
			line++
			col = 0
			continue
		}
		col++
	}

	return line, col
}
