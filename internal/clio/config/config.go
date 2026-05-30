package config

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

type App struct {
	Home string `env:"CLIO_HOME" yaml:"home"`
	File string `env:"CLIO_FILE" yaml:"file"`

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

type fileConfig struct {
	Home                string `yaml:"home"`
	File                string `yaml:"file"`
	Theme               string `yaml:"theme"`
	PrimaryColor        string `yaml:"primary_color"`
	PrimaryColorSubdued string `yaml:"primary_color_subdued"`
	BrightGreenColor    string `yaml:"bright_green"`
	GreenColor          string `yaml:"green"`
	BrightRedColor      string `yaml:"bright_red"`
	RedColor            string `yaml:"red"`
	ForegroundColor     string `yaml:"foreground"`
	BackgroundColor     string `yaml:"background"`
	GrayColor           string `yaml:"gray"`
	BlackColor          string `yaml:"black"`
	WhiteColor          string `yaml:"white"`
}

func defaultApp() App {
	return App{
		Home:                defaultHome(),
		File:                "notes.json",
		Theme:               "materialpalenight",
		PrimaryColor:        "#82AAFF",
		PrimaryColorSubdued: "#676E95",
		BrightGreenColor:    "#C3E88D",
		GreenColor:          "#7FB47C",
		BrightRedColor:      "#FF5370",
		RedColor:            "#C95E78",
		ForegroundColor:     "#A6ACCD",
		BackgroundColor:     "#292D3E",
		GrayColor:           "#676E95",
		BlackColor:          "#292D3E",
		WhiteColor:          "#FFFFFF",
	}
}

func defaultHome() string {
	if home := os.Getenv("CLIO_HOME"); home != "" {
		return home
	}
	return filepath.Join(xdg.DataHome, "clio")
}

func ConfigPath() string {
	if path := os.Getenv("CLIO_CONFIG"); path != "" {
		return path
	}
	cfgPath, err := xdg.ConfigFile("clio/config.yaml")
	if err == nil {
		return cfgPath
	}
	return "config.yaml"
}

func StatePath() string {
	if path := os.Getenv("CLIO_STATE"); path != "" {
		return path
	}
	statePath, err := xdg.StateFile("clio/state.json")
	if err == nil {
		return statePath
	}
	return "state.json"
}

func Load() App {
	app := defaultApp()
	fi, err := os.Open(ConfigPath())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return defaultApp()
	}
	if fi != nil {
		defer fi.Close()
		var cfg fileConfig
		if err := yaml.NewDecoder(fi).Decode(&cfg); err == nil {
			applyFileConfig(&app, cfg)
		}
	}
	if err := env.Parse(&app); err != nil {
		return defaultApp()
	}
	app.Home = expandHome(app.Home)
	return app
}

func (app App) Write() error {
	fi, err := os.Create(ConfigPath())
	if err != nil {
		return err
	}
	defer fi.Close()
	return yaml.NewEncoder(fi).Encode(&app)
}

func applyFileConfig(app *App, cfg fileConfig) {
	if cfg.Home != "" {
		app.Home = cfg.Home
	}
	if cfg.File != "" {
		app.File = cfg.File
	}
	if cfg.Theme != "" {
		app.Theme = cfg.Theme
	}
	if cfg.PrimaryColor != "" {
		app.PrimaryColor = cfg.PrimaryColor
	}
	if cfg.PrimaryColorSubdued != "" {
		app.PrimaryColorSubdued = cfg.PrimaryColorSubdued
	}
	if cfg.BrightGreenColor != "" {
		app.BrightGreenColor = cfg.BrightGreenColor
	}
	if cfg.GreenColor != "" {
		app.GreenColor = cfg.GreenColor
	}
	if cfg.BrightRedColor != "" {
		app.BrightRedColor = cfg.BrightRedColor
	}
	if cfg.RedColor != "" {
		app.RedColor = cfg.RedColor
	}
	if cfg.ForegroundColor != "" {
		app.ForegroundColor = cfg.ForegroundColor
	}
	if cfg.BackgroundColor != "" {
		app.BackgroundColor = cfg.BackgroundColor
	}
	if cfg.GrayColor != "" {
		app.GrayColor = cfg.GrayColor
	}
	if cfg.BlackColor != "" {
		app.BlackColor = cfg.BlackColor
	}
	if cfg.WhiteColor != "" {
		app.WhiteColor = cfg.WhiteColor
	}
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}
