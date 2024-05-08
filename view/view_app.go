package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type App struct {
	optimizationsView *OptimizationsView
	jobs              *JobsView

	width  int
	height int
}

func NewApp(optimizationsView *OptimizationsView, jobs *JobsView) *App {
	r := &App{
		optimizationsView: optimizationsView,
		jobs:              jobs,
	}
	return r
}

func (m *App) Init() tea.Cmd {
	optTableCmd := m.optimizationsView.Init()
	return tea.Batch(optTableCmd, tea.EnterAltScreen)
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.UpdateResponsive()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.jobs.Update(msg)
	_, optTableCmd := m.optimizationsView.Update(msg)
	return m, tea.Batch(optTableCmd)
}

func (m *App) View() string {
	if !m.checkResponsive() {
		m.UpdateResponsive()
		return "Application cannot be rendered in this screen size, please increase height of your terminal"
	}
	sb := strings.Builder{}
	sb.WriteString(m.optimizationsView.View())
	sb.WriteString(m.jobs.View())
	return sb.String()
}

func (m *App) checkResponsive() bool {
	return m.height >= m.jobs.height+m.optimizationsView.height && m.jobs.IsResponsive() && m.optimizationsView.IsResponsive()
}

func (m *App) UpdateResponsive() {
	m.optimizationsView.SetHeight(m.optimizationsView.MinHeight())
	m.jobs.SetHeight(m.jobs.MinHeight())

	if !m.checkResponsive() {
		return // nothing we can do
	}

	for m.optimizationsView.height < m.optimizationsView.PreferredMinHeight() {
		m.optimizationsView.SetHeight(m.optimizationsView.height + 1)
		if !m.checkResponsive() {
			m.optimizationsView.SetHeight(m.optimizationsView.height - 1)
			return
		}
	}

	for m.jobs.height < m.jobs.MaxHeight() {
		m.jobs.SetHeight(m.jobs.height + 1)
		if !m.checkResponsive() {
			m.jobs.SetHeight(m.jobs.height - 1)
			return
		}
	}

	for m.optimizationsView.height < m.optimizationsView.MaxHeight() {
		m.optimizationsView.SetHeight(m.optimizationsView.height + 1)
		if !m.checkResponsive() {
			m.optimizationsView.SetHeight(m.optimizationsView.height - 1)
			return
		}
	}
}
