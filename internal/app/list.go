package app

import (
	"fmt"
	"io"
	"time"

	"github.com/aquilax/truncate"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type noteDelegate struct {
	styles NotesBaseStyle
	state  browseState
}

func (d noteDelegate) Height() int {
	return 2
}

func (d noteDelegate) Spacing() int {
	return 1
}

func (d noteDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return func() tea.Msg {
		if n, ok := m.SelectedItem().(Note); ok {
			return updateContentMsg(n)
		}
		return nil
	}
}

func (d noteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if item == nil {
		return
	}
	n, ok := item.(Note)
	if !ok {
		return
	}

	titleStyle := d.styles.SelectedTitle
	subtitleStyle := d.styles.SelectedSubtitle
	if d.state == copyVisual {
		titleStyle = d.styles.CopiedTitle
		subtitleStyle = d.styles.CopiedSubtitle
	} else if d.state == deleteConfirm {
		titleStyle = d.styles.DeletedTitle
		subtitleStyle = d.styles.DeletedSubtitle
	}

	if index == m.Index() {
		fmt.Fprintln(w, " "+titleStyle.Render("▸ "+truncate.Truncate(n.Title, 28, "...", truncate.PositionEnd)))
		fmt.Fprint(w, " "+subtitleStyle.Render("  "+n.Folder+" • "+humanizeTime(n.Date)))
		return
	}
	fmt.Fprintln(w, " "+d.styles.UnselectedTitle.Render("  "+truncate.Truncate(n.Title, 28, "...", truncate.PositionEnd)))
	fmt.Fprint(w, " "+d.styles.UnselectedSubtitle.Render("  "+n.Folder+" • "+humanizeTime(n.Date)))
}

type folderDelegate struct{ styles FoldersBaseStyle }

func (d folderDelegate) Height() int {
	return 1
}

func (d folderDelegate) Spacing() int {
	return 0
}

func (d folderDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d folderDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	f, ok := item.(Folder)
	if !ok {
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, d.styles.Selected.Render("▸ "+string(f)))
		return
	}
	fmt.Fprint(w, d.styles.Unselected.Render("  "+string(f)))
}

const (
	Day   = 24 * time.Hour
	Week  = 7 * Day
	Month = 30 * Day
	Year  = 12 * Month
)

var magnitudes = []humanize.RelTimeMagnitude{
	{D: 5 * time.Second, Format: "just now", DivBy: time.Second},
	{D: time.Minute, Format: "moments ago", DivBy: time.Second},
	{D: time.Hour, Format: "%dm %s", DivBy: time.Minute},
	{D: 2 * time.Hour, Format: "1h %s", DivBy: 1},
	{D: Day, Format: "%dh %s", DivBy: time.Hour},
	{D: 2 * Day, Format: "1d %s", DivBy: 1},
	{D: Week, Format: "%dd %s", DivBy: Day},
	{D: 2 * Week, Format: "1w %s", DivBy: 1},
	{D: Month, Format: "%dw %s", DivBy: Week},
	{D: 2 * Month, Format: "1mo %s", DivBy: 1},
	{D: Year, Format: "%dmo %s", DivBy: Month},
	{D: 18 * Month, Format: "1y %s", DivBy: 1},
	{D: 2 * Year, Format: "2y %s", DivBy: 1},
}

func humanizeTime(t time.Time) string {
	return humanize.CustomRelTime(t, time.Now(), "ago", "from now", magnitudes)
}
