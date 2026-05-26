package app

import "github.com/charmbracelet/lipgloss"

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

var (
	PrimaryColor   = lipgloss.Color("#7aa2f7")
	SecondaryColor = lipgloss.Color("#bb9af7")
	SuccessColor   = lipgloss.Color("#9ece6a")
	WarningColor   = lipgloss.Color("#e0af68")
	AlertColor     = lipgloss.Color("#f7768e")
	BgColor        = lipgloss.Color("#1a1b26")
	SelectionColor = lipgloss.Color("#283457")
	MutedColor     = lipgloss.Color("#565f89")
	FgColor        = lipgloss.Color("#a9b1d6")
	WhiteColor     = lipgloss.Color("#c0caf5")
	BlackColor     = lipgloss.Color("#24283b")
	BorderColor    = lipgloss.Color("#3b4261")

	TitleStyle = lipgloss.NewStyle().
			Foreground(BgColor).
			Background(PrimaryColor).
			Bold(true).
			Padding(0, 1)

	AccentTitleStyle = lipgloss.NewStyle().
				Foreground(BgColor).
				Background(SecondaryColor).
				Bold(true).
				Padding(0, 1)

	ModeStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	AlertStyle = lipgloss.NewStyle().
			Foreground(AlertColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	BorderedBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1).
			Background(BgColor)

	ActiveBorderedBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Padding(0, 1).
				Background(BgColor)

	ModalBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2).
			Align(lipgloss.Center).
			Background(BgColor)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(WhiteColor).
			Bold(true).
			Background(PrimaryColor).
			Padding(0, 1).
			MarginRight(1)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Background(BgColor).
			Padding(0, 0)

	PanelActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Background(BgColor).
				Padding(0, 0)

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Bold(true).
			Padding(0, 1)

	PanelTitleActiveStyle = lipgloss.NewStyle().
				Foreground(WhiteColor).
				Bold(true).
				Background(PrimaryColor).
				Padding(0, 1).
				MarginBottom(1)

	SelectionStyle = lipgloss.NewStyle().
			Background(SelectionColor).
			Foreground(WhiteColor).
			Bold(true)

	DividerStyle = lipgloss.NewStyle().
			Foreground(BorderColor)
)

func DefaultStyles(config Config) Styles {
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
	selection := lipgloss.Color("#283457")

	return Styles{
		Notes: NotesStyle{
			Focused: NotesBaseStyle{
				Base:               lipgloss.NewStyle().Width(33),
				TitleBar:           lipgloss.NewStyle().Background(blue).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white).Bold(true),
				SelectedSubtitle:   lipgloss.NewStyle().Foreground(gray),
				UnselectedSubtitle: lipgloss.NewStyle().Foreground(brightBlack),
				SelectedTitle:      lipgloss.NewStyle().Foreground(brightBlue).Background(selection).Bold(true),
				UnselectedTitle:    lipgloss.NewStyle().Foreground(white),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white).Bold(true),
				CopiedTitle:        lipgloss.NewStyle().Foreground(brightGreen).Background(selection).Bold(true),
				CopiedSubtitle:     lipgloss.NewStyle().Foreground(green),
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white).Bold(true),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed).Background(selection).Bold(true),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
			Blurred: NotesBaseStyle{
				Base:               lipgloss.NewStyle().Width(33),
				TitleBar:           lipgloss.NewStyle().Background(black).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(gray).Bold(true),
				SelectedSubtitle:   lipgloss.NewStyle().Foreground(blue),
				UnselectedSubtitle: lipgloss.NewStyle().Foreground(black),
				SelectedTitle:      lipgloss.NewStyle().Foreground(brightBlue).Background(selection).Bold(true),
				UnselectedTitle:    lipgloss.NewStyle().Foreground(gray),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Bold(true),
				CopiedTitle:        lipgloss.NewStyle().Foreground(brightGreen).Background(selection).Bold(true),
				CopiedSubtitle:     lipgloss.NewStyle().Foreground(green),
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(33-2).Margin(0, 1, 1, 1).Padding(0, 1).Bold(true),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed).Background(selection).Bold(true),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
		},
		Folders: FoldersStyle{
			Focused: FoldersBaseStyle{
				Base:       lipgloss.NewStyle().Width(20),
				Title:      lipgloss.NewStyle().Padding(0, 1).Foreground(white).Bold(true),
				TitleBar:   lipgloss.NewStyle().Background(blue).Width(20-2).Margin(0, 1, 1, 1).Bold(true),
				Selected:   lipgloss.NewStyle().Foreground(brightBlue).Background(selection).Bold(true),
				Unselected: lipgloss.NewStyle().Foreground(gray),
			},
			Blurred: FoldersBaseStyle{
				Base:       lipgloss.NewStyle().Width(20),
				Title:      lipgloss.NewStyle().Padding(0, 1).Foreground(gray).Bold(true),
				TitleBar:   lipgloss.NewStyle().Background(black).Width(20-2).Margin(0, 1, 1, 1).Bold(true),
				Selected:   lipgloss.NewStyle().Foreground(brightBlue).Background(selection).Bold(true),
				Unselected: lipgloss.NewStyle().Foreground(brightBlack),
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Base:         lipgloss.NewStyle().Margin(0, 1),
				Title:        lipgloss.NewStyle().Background(blue).Foreground(white).Margin(0, 0, 1, 1).Padding(0, 1).Bold(true),
				Separator:    lipgloss.NewStyle().Foreground(white).Margin(0, 0, 1, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(brightBlack),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue).Bold(true),
			},
			Blurred: ContentBaseStyle{
				Base:         lipgloss.NewStyle().Margin(0, 1),
				Title:        lipgloss.NewStyle().Background(black).Foreground(gray).Margin(0, 0, 1, 1).Padding(0, 1).Bold(true),
				Separator:    lipgloss.NewStyle().Foreground(gray).Margin(0, 0, 1, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(black),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue).Bold(true),
			},
		},
	}
}
