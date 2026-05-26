package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Home string `env:"CLIO_HOME" yaml:"home"`
	File string `env:"CLIO_FILE" yaml:"file"`

	DefaultLanguage string `env:"CLIO_DEFAULT_LANGUAGE" yaml:"default_language"`

	Theme string `env:"CLIO_THEME" yaml:"theme"`

	PrimaryColor        string `env:"CLIO_PRIMARY_COLOR" yaml:"primary_color"`
	PrimaryColorSubdued string `env:"CLIO_PRIMARY_COLOR_SUBDUED" yaml:"primary_color_subdued"`
	BrightGreenColor    string `env:"CLIO_BRIGHT_GREEN" yaml:"bright_green"`
	GreenColor          string `env:"CLIO_GREEN" yaml:"green"`
	BrightRedColor      string `env:"CLIO_BRIGHT_RED" yaml:"bright_red"`
	RedColor            string `env:"CLIO_RED" yaml:"red"`
	ForegroundColor     string `env:"CLIO_FOREGROUND" yaml:"foreground"`
	BackgroundColor     string `env:"CLIO_BACKGROUND" yaml:"background"`
	GrayColor           string `env:"CLIO_GRAY" yaml:"gray"`
	BlackColor          string `env:"CLIO_BLACK" yaml:"black"`
	WhiteColor          string `env:"CLIO_WHITE" yaml:"white"`
}

func newConfig() Config {
	return Config{
		Home:                defaultHome(),
		File:                "notes",
		DefaultLanguage:     "markdown",
		Theme:               "tokyonight",
		PrimaryColor:        "#7aa2f7",
		PrimaryColorSubdued: "#3b4261",
		BrightGreenColor:    "#9ece6a",
		GreenColor:          "#73daca",
		BrightRedColor:      "#f7768e",
		RedColor:            "#db4b4b",
		ForegroundColor:     "#a9b1d6",
		BackgroundColor:     "#1a1b26",
		GrayColor:           "#565f89",
		BlackColor:          "#24283b",
		WhiteColor:          "#c0caf5",
	}
}

func defaultHome() string { return filepath.Join(xdg.DataHome, "clio") }

func defaultConfig() string {
	if c := os.Getenv("CLIO_CONFIG"); c != "" {
		return c
	}
	cfgPath, err := xdg.ConfigFile("clio/config.yaml")
	if err != nil {
		return "config.yaml"
	}
	return cfgPath
}

func ReadConfig() Config {
	config := newConfig()
	fi, err := os.Open(defaultConfig())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return newConfig()
	}
	if fi != nil {
		defer fi.Close()
		if err := yaml.NewDecoder(fi).Decode(&config); err != nil {
			return newConfig()
		}
	}

	if err := env.Parse(&config); err != nil {
		return newConfig()
	}

	if strings.HasPrefix(config.Home, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			config.Home = filepath.Join(home, config.Home[1:])
		}
	}

	return config
}
