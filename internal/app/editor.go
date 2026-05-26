package app

import (
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "nano"

func editorCmd(path string) *exec.Cmd {
	editor, args := getEditor()
	return exec.Command(editor, append(args, path)...)
}

func getEditor() (string, []string) {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) > 0 {
		return editor[0], editor[1:]
	}
	return defaultEditor, nil
}
