package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kaytu-io/kaytu/pkg/metrics"
	"github.com/kaytu-io/kaytu/pkg/provider"
	"strings"
)

type App struct {
	provider       provider.Provider
	metricProvider metrics.MetricProvider
	identification map[string]string

	processInstanceChan chan OptimizationItem
	optimizationsTable  *Ec2InstanceOptimizations
	jobs                *JobsView

	width  int
	height int
}

var (
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func NewApp(prv provider.Provider, metric metrics.MetricProvider, identification map[string]string) *App {
	pi := make(chan OptimizationItem, 1000)
	r := &App{
		processInstanceChan: pi,
		optimizationsTable:  NewEC2InstanceOptimizations(pi),
		jobs:                NewJobsView(),
		provider:            prv,
		metricProvider:      metric,
		identification:      identification,
	}
	go r.ProcessWastages()
	go r.ProcessAllRegions()
	return r
}

func (m *App) Init() tea.Cmd {
	optTableCmd := m.optimizationsTable.Init()
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
	_, optTableCmd := m.optimizationsTable.Update(msg)
	return m, tea.Batch(optTableCmd)
}

func (m *App) View() string {
	if !m.checkResponsive() {
		return "Application cannot be rendered in this screen size, please increase height of your terminal"
	}
	sb := strings.Builder{}
	sb.WriteString(m.optimizationsTable.View())
	sb.WriteString(m.jobs.View())
	return sb.String()
}

func (m *App) checkResponsive() bool {
	return m.height >= m.jobs.height+m.optimizationsTable.height && m.jobs.IsResponsive() && m.optimizationsTable.IsResponsive()
}

func (m *App) UpdateResponsive() {
	m.optimizationsTable.SetHeight(m.optimizationsTable.MinHeight())
	m.jobs.SetHeight(m.jobs.MinHeight())

	if !m.checkResponsive() {
		return // nothing we can do
	}

	for m.optimizationsTable.height < m.optimizationsTable.PreferredMinHeight() {
		m.optimizationsTable.SetHeight(m.optimizationsTable.height + 1)
		if !m.checkResponsive() {
			m.optimizationsTable.SetHeight(m.optimizationsTable.height - 1)
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

	for m.optimizationsTable.height < m.optimizationsTable.MaxHeight() {
		m.optimizationsTable.SetHeight(m.optimizationsTable.height + 1)
		if !m.checkResponsive() {
			m.optimizationsTable.SetHeight(m.optimizationsTable.height - 1)
			return
		}
	}
}
