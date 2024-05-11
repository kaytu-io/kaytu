package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type tickMsg time.Time

func TickCmdWithDuration(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
