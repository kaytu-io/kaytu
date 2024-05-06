package plugin

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view"
	"github.com/spf13/cobra"
	"time"
)

type App struct {
	table       table.Model
	renderCount int
}

func NewApp() (*App, error) {
	plugins, err := server.GetPlugins()
	if err != nil {
		return nil, err
	}

	columns := []table.Column{
		table.NewColumn("name", "Name", 30),
		table.NewColumn("version", "Version", 30),
		table.NewColumn("provider", "Provider", 30),
	}

	var rows []table.Row
	for _, plg := range plugins {
		rows = append(rows, table.Row{
			Data: table.RowData{
				"name":     plg.Config.Name,
				"version":  plg.Config.Version,
				"provider": plg.Config.Provider,
			},
		})
	}

	return &App{
		table: table.New(columns).
			WithRows(rows).
			WithFooterVisibility(false).
			WithBaseStyle(style.Base).
			BorderRounded(),
	}, nil
}

func (m *App) Init() tea.Cmd {
	return tea.Batch(m.table.Init(), view.TickCmdWithDuration(time.Millisecond*50))
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.table, cmd = m.table.Update(msg)
	if m.renderCount > 10 {
		return m, tea.Quit
	}

	return m, tea.Batch(cmd, view.TickCmdWithDuration(time.Millisecond*50))
}

func (m *App) View() string {
	m.renderCount++
	return m.table.View() + "\n"
}

var listCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {

		app, err := NewApp()
		if err != nil {
			return err
		}

		p := tea.NewProgram(app)
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}
