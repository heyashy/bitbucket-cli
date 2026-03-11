package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colors
	Blue   = lipgloss.Color("#0052CC")
	Green  = lipgloss.Color("#36B37E")
	Red    = lipgloss.Color("#FF5630")
	Yellow = lipgloss.Color("#FFAB00")
	Purple = lipgloss.Color("#6554C0")
	Grey   = lipgloss.Color("#97A0AF")
	White  = lipgloss.Color("#FFFFFF")
	Subtle = lipgloss.Color("#6B778C")

	// Text styles
	Bold      = lipgloss.NewStyle().Bold(true)
	Faint     = lipgloss.NewStyle().Foreground(Subtle)
	Success   = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Error     = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Warning   = lipgloss.NewStyle().Foreground(Yellow).Bold(true)
	Info      = lipgloss.NewStyle().Foreground(Blue).Bold(true)
	Highlight = lipgloss.NewStyle().Foreground(Purple).Bold(true)

	// Status badges
	BadgeOpen     = lipgloss.NewStyle().Background(Blue).Foreground(White).Padding(0, 1).Bold(true)
	BadgeMerged   = lipgloss.NewStyle().Background(Purple).Foreground(White).Padding(0, 1).Bold(true)
	BadgeDeclined = lipgloss.NewStyle().Background(Red).Foreground(White).Padding(0, 1).Bold(true)
	BadgeApproved = lipgloss.NewStyle().Background(Green).Foreground(White).Padding(0, 1).Bold(true)

	// Sections
	Title = lipgloss.NewStyle().Bold(true).Foreground(White).MarginBottom(1)

	// Borders
	Divider = lipgloss.NewStyle().Foreground(Grey).SetString("─────────────────────────────────────────")

	// PR list row
	PRNumber = lipgloss.NewStyle().Foreground(Blue).Bold(true).Width(6)
	PRTitle  = lipgloss.NewStyle().Bold(true).Width(50)
	PRBranch = lipgloss.NewStyle().Foreground(Subtle)
	PRAuthor = lipgloss.NewStyle().Foreground(Purple)

	// Key-value display
	Key   = lipgloss.NewStyle().Foreground(Grey).Width(12)
	Value = lipgloss.NewStyle().Bold(true)
)

func StatusBadge(state string) string {
	switch state {
	case "OPEN":
		return BadgeOpen.Render("OPEN")
	case "MERGED":
		return BadgeMerged.Render("MERGED")
	case "DECLINED":
		return BadgeDeclined.Render("DECLINED")
	case "SUPERSEDED":
		return BadgeDeclined.Render("SUPERSEDED")
	default:
		return Faint.Render(state)
	}
}

func CheckMark() string {
	return Success.Render("✓")
}

func CrossMark() string {
	return Error.Render("✗")
}

func Arrow() string {
	return Faint.Render("→")
}

func Spinner() string {
	return Info.Render("●")
}
