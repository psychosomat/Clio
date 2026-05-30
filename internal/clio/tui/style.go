package tui

import (
	appconfig "clio/internal/clio/config"

	"github.com/charmbracelet/lipgloss"
)

type NotesStyle struct {
	Focused NotesBaseStyle
	Blurred NotesBaseStyle
}

type FoldersStyle struct {
	Focused FoldersBaseStyle
	Blurred FoldersBaseStyle
}

type ContentStyle struct {
	Focused ContentBaseStyle
	Blurred ContentBaseStyle
}

type NotesBaseStyle struct {
	Base               lipgloss.Style
	Title              lipgloss.Style
	TitleBar           lipgloss.Style
	SelectedSubtitle   lipgloss.Style
	UnselectedSubtitle lipgloss.Style
	SelectedTitle      lipgloss.Style
	UnselectedTitle    lipgloss.Style
	CopiedTitleBar     lipgloss.Style
	CopiedTitle        lipgloss.Style
	CopiedSubtitle     lipgloss.Style
	DeletedTitleBar    lipgloss.Style
	DeletedTitle       lipgloss.Style
	DeletedSubtitle    lipgloss.Style
}

type FoldersBaseStyle struct {
	Base       lipgloss.Style
	Title      lipgloss.Style
	TitleBar   lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
}

type ContentBaseStyle struct {
	Base         lipgloss.Style
	Title        lipgloss.Style
	Separator    lipgloss.Style
	LineNumber   lipgloss.Style
	EmptyHint    lipgloss.Style
	EmptyHintKey lipgloss.Style
}

type Styles struct {
	Notes   NotesStyle
	Folders FoldersStyle
	Content ContentStyle
}

var marginStyle = lipgloss.NewStyle().Margin(1, 0, 0, 1)

func DefaultStyles(config appconfig.App) Styles {
	white := lipgloss.Color(config.WhiteColor)
	gray := lipgloss.Color(config.GrayColor)
	black := lipgloss.Color(config.BackgroundColor)
	brightBlack := lipgloss.Color(config.BlackColor)
	green := lipgloss.Color(config.GreenColor)
	brightGreen := lipgloss.Color(config.BrightGreenColor)
	brightBlue := lipgloss.Color(config.PrimaryColor)
	blue := lipgloss.Color(config.PrimaryColorSubdued)
	red := lipgloss.Color(config.RedColor)
	brightRed := lipgloss.Color(config.BrightRedColor)

	return Styles{
		Notes: NotesStyle{
			Focused: NotesBaseStyle{
				Base:               lipgloss.NewStyle().Width(35),
				TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				SelectedSubtitle:   lipgloss.NewStyle().Foreground(blue),
				UnselectedSubtitle: lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
				SelectedTitle:      lipgloss.NewStyle().Foreground(brightBlue),
				UnselectedTitle:    lipgloss.NewStyle().Foreground(gray),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				CopiedTitle:        lipgloss.NewStyle().Foreground(brightGreen),
				CopiedSubtitle:     lipgloss.NewStyle().Foreground(green),
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
			Blurred: NotesBaseStyle{
				Base:               lipgloss.NewStyle().Width(35),
				TitleBar:           lipgloss.NewStyle().Background(black).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(gray),
				SelectedSubtitle:   lipgloss.NewStyle().Foreground(blue),
				UnselectedSubtitle: lipgloss.NewStyle().Foreground(black),
				SelectedTitle:      lipgloss.NewStyle().Foreground(brightBlue),
				UnselectedTitle:    lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1),
				CopiedTitle:        lipgloss.NewStyle().Foreground(brightGreen),
				CopiedSubtitle:     lipgloss.NewStyle().Foreground(green),
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
		},
		Folders: FoldersStyle{
			Focused: FoldersBaseStyle{
				Base:       lipgloss.NewStyle().Width(22),
				Title:      lipgloss.NewStyle().Padding(0, 1).Foreground(white),
				TitleBar:   lipgloss.NewStyle().Background(blue).Width(22-2).Margin(0, 1, 1, 1),
				Selected:   lipgloss.NewStyle().Foreground(brightBlue),
				Unselected: lipgloss.NewStyle().Foreground(gray),
			},
			Blurred: FoldersBaseStyle{
				Base:       lipgloss.NewStyle().Width(22),
				Title:      lipgloss.NewStyle().Padding(0, 1).Foreground(gray),
				TitleBar:   lipgloss.NewStyle().Background(black).Width(22-2).Margin(0, 1, 1, 1),
				Selected:   lipgloss.NewStyle().Foreground(brightBlue),
				Unselected: lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Base:         lipgloss.NewStyle().Margin(0, 1),
				Title:        lipgloss.NewStyle().Background(blue).Foreground(white).Margin(0, 0, 1, 1).Padding(0, 1),
				Separator:    lipgloss.NewStyle().Foreground(white).Margin(0, 0, 1, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(brightBlack),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue),
			},
			Blurred: ContentBaseStyle{
				Base:         lipgloss.NewStyle().Margin(0, 1),
				Title:        lipgloss.NewStyle().Background(black).Foreground(gray).Margin(0, 0, 1, 1).Padding(0, 1),
				Separator:    lipgloss.NewStyle().Foreground(gray).Margin(0, 0, 1, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(black),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue),
			},
		},
	}
}
