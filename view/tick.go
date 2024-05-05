package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return TickCmdWithDuration(time.Second)
}

func TickCmdWithDuration(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
