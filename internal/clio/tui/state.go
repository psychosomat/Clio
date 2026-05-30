package tui

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	appconfig "clio/internal/clio/config"
)

type State struct {
	CurrentFolder string
	CurrentNote   string
}

func (s State) Save() error {
	fi, err := os.Create(defaultState())
	if err != nil {
		return err
	}
	defer fi.Close()
	return json.NewEncoder(fi).Encode(s)
}

func defaultState() string {
	return appconfig.StatePath()
}

func readState() State {
	var s State
	fi, err := os.Open(defaultState())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return s
	}
	defer fi.Close()

	if err := json.NewDecoder(fi).Decode(&s); err != nil {
		return s
	}

	return s
}
