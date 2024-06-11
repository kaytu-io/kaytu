package style

import (
	"github.com/charmbracelet/lipgloss"
	"regexp"
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	HelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	Bold       = lipgloss.NewStyle().Bold(true)
	ChangeFrom = lipgloss.NewStyle().Background(lipgloss.Color("88")).Foreground(lipgloss.Color("#ffffff"))
	ChangeTo   = lipgloss.NewStyle().Background(lipgloss.Color("28")).Foreground(lipgloss.Color("#ffffff"))
	Base       = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("238")).
			Align(lipgloss.Left)
	ActiveStyleBase = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("248")).
			Align(lipgloss.Left)
	CostStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	SavingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	InputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	ContinueStyle = lipgloss.NewStyle().Foreground(darkGray)
	SvcDisable    = lipgloss.NewStyle().Background(lipgloss.Color("#222222"))
	SvcEnable     = lipgloss.NewStyle().Background(lipgloss.Color("#aa2222"))

	StatusBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#222222")).Foreground(lipgloss.Color("#ffffff")).Width(9999)
	JobsStatusStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#dd5200")).Foreground(lipgloss.Color("#ffffff"))
	ErrorStatusStyle = lipgloss.NewStyle().Background(lipgloss.Color("#aa2222")).Foreground(lipgloss.Color("#ffffff"))

	HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6e6f4")).Background(lipgloss.Color("#3a5369"))

	InfoStatusStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#3a3835")).Foreground(lipgloss.Color("#ffffff"))
	InfoStatusStyle2 = lipgloss.NewStyle().Background(lipgloss.Color("#006d69")).Foreground(lipgloss.Color("#ffffff"))
)

var StyleSelector = regexp.MustCompile(`\x1b\[[0-9;]*m`)
