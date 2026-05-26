![preview](./preview.png)

<p align="center">
  <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='120' viewBox='0 0 400 120'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0%25' y1='0%25' x2='100%25' y2='100%25'%3E%3Cstop offset='0%25' style='stop-color:%23bd93f9'/%3E%3Cstop offset='100%25' style='stop-color:%23ff79c6'/%3E%3C/linearGradient%3E%3C/defs%3E%3Ctext x='200' y='72' font-family='ui-monospace,SFMono-Regular,SF Mono,Menlo,Consolas,monospace' font-size='64' font-weight='900' text-anchor='middle' fill='url(%23g)'%3EClio%3C/text%3E%3Ctext x='200' y='98' font-family='system-ui,-apple-system,sans-serif' font-size='14' fill='%236278a4' text-anchor='middle'%3Ekeyboard-first terminal notes inbox%3C/text%3E%3C/svg%3E" alt="Clio">
</p>

<p align="center">
  <a href="https://github.com/psychosomat/Clio/releases"><img src="https://img.shields.io/github/v/release/psychosomat/Clio?style=flat-square&label=release&color=%23bd93f9" alt="Release"></a>
  <a href="https://aur.archlinux.org/packages/clio"><img src="https://img.shields.io/aur/version/clio?style=flat-square&label=AUR&color=%231793d1" alt="AUR"></a>
  <a href="https://github.com/psychosomat/Clio/blob/main/LICENSE"><img src="https://img.shields.io/github/license/psychosomat/Clio?style=flat-square&color=%23ff79c6" alt="License"></a>
</p>

**Clio** is a keyboard-driven TUI notes app. Notes are plain Markdown files with YAML front matter — no database, no lock-in. Multi-pane interface with folders, live search, Markdown preview, and autosave.

## Features

- **Multi-pane** — folders, notes list, and rendered Markdown preview side by side
- **Markdown editor** — full editor with line numbers, snippet insertion, and autosave
- **Rich preview** — syntax-highlighted code blocks, Mermaid diagrams, math, wiki links, callouts
- **Live search** — filters notes by title and body as you type
- **Folders** — organize notes into virtual folders stored in front matter
- **Archive** — toggle visibility of archived notes without deleting them
- **Safe trash** — notes are moved to `trash/`, never permanently deleted
- **Copy & paste** — duplicate notes within the app
- **External editor** — open notes in `$EDITOR` or `nano`
- **Reorder** — move notes up/down with `j`/`k`
- **Tokyo Night** — default theme with full color customization via config or env vars

## Installation

| Package | Command |
|---------|---------|
| AUR | `yay -S clio` |
| Pacman | Download `.pkg.tar.zst` from [releases](https://github.com/psychosomat/Clio/releases) |
| Debian/Ubuntu | Download `.deb` from [releases](https://github.com/psychosomat/Clio/releases) |
| Source | see [Development](#development) |

## Usage

```
clio              Launch interactive TUI
clio new          Create a new note in editor mode
clio list         List all note titles (non-interactive)
clio <query>      Fuzzy-find a note and print its body
clio -h, --help   Show help
echo "text" | clio [title]   Save piped input as a new note
```

### Browsing

| Key | Action |
|-----|--------|
| `q`, `ctrl+c` | Quit |
| `?` | Toggle help overlay |
| `/` | Search / filter |
| `n` | New note |
| `e` | Edit selected note |
| `Enter` | Open selected note in editor |
| `x` | Delete (move to trash) |
| `c` | Copy note |
| `p` | Paste note |
| `r` | Rename note |
| `R` | Move to folder |
| `N` | Create folder |
| `a` | Archive / unarchive |
| `A` | Toggle show archived |
| `j`, `k` | Move note down / up (reorder) |
| `Tab`, `→` | Next pane |
| `Shift+Tab`, `←` | Previous pane |
| `F2` | Toggle preview pane |

### Editing

| Key | Action |
|-----|--------|
| `Esc` | Back to browsing (autosave) |
| `F2` | Toggle Markdown preview |
| `Ctrl+E` | Open in external editor |
| `F3` | Insert code block |
| `F4` | Insert table |
| `F5` | Insert checklist |
| `F6` | Insert quote |
| `F7` | Insert link |
| `F8` | Insert heading |
| `F9` | Insert horizontal rule |

## Configuration

Config file at `~/.config/clio/config.yaml` (YAML):

```yaml
theme: tokyonight
primary_color: "#bd93f9"
default_language: text
```

All settings can be overridden with environment variables: `CLIO_THEME`, `CLIO_HOME`, `CLIO_PRIMARY_COLOR`, `CLIO_PRIMARY_COLOR_SUBDUED`, `CLIO_BRIGHT_GREEN`, `CLIO_GREEN`, `CLIO_BRIGHT_RED`, `CLIO_RED`, `CLIO_FOREGROUND`, `CLIO_BACKGROUND`, `CLIO_GRAY`, `CLIO_BLACK`, `CLIO_WHITE`, `CLIO_DEFAULT_LANGUAGE`.

## Storage

Files live under the [XDG](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) data directory:

```
~/.local/share/clio/
├── notes/     # Active notes
└── trash/     # Trashed notes
```

Session state is persisted at `~/.local/state/clio/state.json`.

Each note is a timestamped Markdown file with YAML front matter:

```markdown
---
id: "20260526-001122-example-title"
created_at: 2026-05-26T00:11:22+03:00
updated_at: 2026-05-26T00:15:04+03:00
position: 0
archived: false
folder: notes
---

Content starts here.
```

Being plain Markdown, notes work with any editor or sync service (Dropbox, Nextcloud, Git).

## Development

Requirements: [Go](https://go.dev/dl/) 1.26.3+.

```bash
git clone https://github.com/psychosomat/Clio.git
cd Clio
go run ./cmd/clio
```

Tests:

```bash
go test ./internal/...              # all tests
go test ./internal/notes/...        # storage and search
go test ./internal/app/...          # editor, preview, and view
go test ./internal/markdownpreview/ # markdown renderer
```

### Project layout

```
├── cmd/clio                  Entrypoint, flags, config
├── internal/notes            Domain model, YAML parsing, search, storage
├── internal/app              Bubble Tea model/update/view, editor, keymap, theme
├── internal/markdownpreview  Goldmark-based Markdown → ANSI renderer
└── scripts/                  Release, packaging, and AUR helpers
```

## License

[MIT](./LICENSE)
