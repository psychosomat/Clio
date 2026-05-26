![preview](./preview.png)

<p align="center">
  <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='120' viewBox='0 0 400 120'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0%25' y1='0%25' x2='100%25' y2='100%25'%3E%3Cstop offset='0%25' style='stop-color:%23bd93f9'/%3E%3Cstop offset='100%25' style='stop-color:%23ff79c6'/%3E%3C/linearGradient%3E%3C/defs%3E%3Ctext x='200' y='72' font-family='ui-monospace,SFMono-Regular,SF Mono,Menlo,Consolas,monospace' font-size='64' font-weight='900' text-anchor='middle' fill='url(%23g)'%3EClio%3C/text%3E%3Ctext x='200' y='98' font-family='system-ui,-apple-system,sans-serif' font-size='14' fill='%236278a4' text-anchor='middle'%3Ekeyboard-first terminal notes inbox%3C/text%3E%3C/svg%3E" alt="Clio">
</p>

<p align="center">
  <a href="https://github.com/psychosomat/Clio/releases"><img src="https://img.shields.io/github/v/release/psychosomat/Clio?style=flat-square&label=release&color=%23bd93f9" alt="Release"></a>
  <a href="https://aur.archlinux.org/packages/clio"><img src="https://img.shields.io/aur/version/clio?style=flat-square&label=AUR&color=%231793d1" alt="AUR"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-ff79c6?style=flat-square" alt="MIT"></a>
</p>

**Clio** is a TUI notes inbox built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Notes are plain Markdown files with YAML front matter — no database, no lock-in, keyboard-only.

## Features

- **Instant** — loads hundreds of notes in milliseconds
- **Portable** — all notes are plain `.md` files, editable by any tool
- **Live search** — filters titles and bodies as you type
- **Dracula theme** — polished UI with side-by-side preview
- **Safe trash** — notes are renamed and moved, never permanently deleted
- **Autosave** — saves automatically after a pause in typing

## Installation

| Package | Command |
|---------|---------|
| AUR | `yay -S clio` |
| Debian/Ubuntu | Download `.deb` from [releases](https://github.com/psychosomat/Clio/releases) |
| Go | `go install github.com/psychosomat/Clio/cmd/clio@latest` |
| Source | see [Development](#development) |

## Usage

```
clio            Open inbox
clio new        Create a note immediately
clio list       Open inbox (default)
clio --help     Show help
clio --dir      Custom notes directory
```

| Key | Action |
|-----|--------|
| `n` | New note |
| `Enter` | Open note |
| `a` | Archive / unarchive |
| `d` | Move to trash |
| `/` | Search |
| `q` | Quit |
| `j`/`k`, `↑`/`↓` | Navigate |
| `F2` | Toggle preview |
| `F3`–`F9` | Insert code, table, checklist, quote, link, heading, rule |
| `Esc` | Back to inbox |

## Storage

Files live under the [XDG](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) data directory:

```
~/.local/share/clio/
├── notes/     # Active notes
└── trash/     # Trashed notes
```

Each note is a timestamped Markdown file with YAML front matter:

```markdown
---
id: "20260526-001122-example-title"
created_at: "2026-05-26T00:11:22+03:00"
updated_at: "2026-05-26T00:15:04+03:00"
archived: false
---

Content starts here.
```

Being plain Markdown, notes work with any editor or sync service (Dropbox, Nextcloud, Git).

## Development

Requirements: [Go](https://go.dev/dl/) (see `go.mod` for minimum version).

```bash
git clone https://github.com/psychosomat/Clio.git
cd Clio
go run ./cmd/clio  # run it
```

Tests span the storage, search, and editor layers:

```bash
go test ./internal/...         # all internal tests
go test ./internal/notes/...   # storage and search
go test ./internal/app/...     # editor and view
```

### Project layout

```
├── cmd/clio            Entrypoint, flags, config
├── internal/notes      Domain model, YAML parsing, search
├── internal/app        Bubble Tea model/update/view, styling
└── scripts/            Release and packaging helpers
```

## License

[MIT](LICENSE)
