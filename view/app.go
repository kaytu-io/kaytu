package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/view/responsive"
	"time"
)

type PageEnum int

const (
	Page_Overview        = 0
	Page_ResourceDetails = 1
	Page_Preferences     = 2
	Page_Jobs            = 3
	Page_ContactUs       = 4
)

type Page interface {
	OnOpen() Page
	OnClose() Page
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetResponsiveView(rv responsive.ResponsiveViewInterface) Page

	responsive.ResponsiveViewInterface
}

type App struct {
	pages         []Page
	history       []PageEnum
	activePageIdx int
	width, height int
	ignoreESC     bool
}

func NewApp(
	optimizationsPage OverviewPage,
	optimizationDetailsPage ResourceDetailsPage,
	preferencesPage PreferencesPage[golang.OptimizationItem],
	jobsPage JobsPage,
	contactUsPage ContactUsPage,
) *App {
	app := &App{}
	optimizationsPage = optimizationsPage.SetApp(app)
	optimizationDetailsPage = optimizationDetailsPage.SetApp(app)
	app.pages = []Page{
		optimizationsPage,
		optimizationDetailsPage,
		preferencesPage,
		jobsPage,
		contactUsPage,
	}
	return app
}

func NewCustomPluginApp(
	optimizationsPage *PluginCustomOverviewPage,
	optimizationDetailsPage *PluginCustomResourceDetailsPage,
	preferencesPage PreferencesPage[golang.ChartOptimizationItem],
	jobsPage JobsPage,
	contactUsPage ContactUsPage,
) *App {
	app := &App{}
	optimizationsPage = optimizationsPage.SetApp(app)
	optimizationDetailsPage = optimizationDetailsPage.SetApp(app)
	app.pages = []Page{
		optimizationsPage,
		optimizationDetailsPage,
		preferencesPage,
		jobsPage,
		contactUsPage,
	}
	return app
}

func (m *App) ChangePage(id PageEnum) tea.Cmd {
	m.history = append(m.history, PageEnum(m.activePageIdx))
	m.pages[m.activePageIdx] = m.pages[m.activePageIdx].OnClose()

	m.activePageIdx = int(id)
	m.pages[m.activePageIdx] = m.pages[m.activePageIdx].OnOpen()

	wsMsg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	}
	model, updateSizeCmd := m.pages[m.activePageIdx].Update(wsMsg)
	m.pages[m.activePageIdx] = model.(Page)

	newRV := m.pages[m.activePageIdx].SetSize(wsMsg)
	m.pages[m.activePageIdx] = m.pages[m.activePageIdx].SetResponsiveView(newRV)
	updateSizeCmd = tea.Batch(tea.ClearScreen, updateSizeCmd)

	return updateSizeCmd
}

func (m *App) Init() tea.Cmd {
	m.ChangePage(Page_Overview)
	return tea.Batch(m.pages[m.activePageIdx].Init(), tea.EnterAltScreen, TickCmdWithDuration(100*time.Microsecond))
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var changePageCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Height < 20 {
			msg.Height = 20
		}
		m.height = msg.Height
		m.width = msg.Width

		model, updateSizeCmd := m.pages[m.activePageIdx].Update(msg)
		m.pages[m.activePageIdx] = model.(Page)
		changePageCmd = updateSizeCmd

		newRV := m.pages[m.activePageIdx].SetSize(msg)
		m.pages[m.activePageIdx] = m.pages[m.activePageIdx].SetResponsiveView(newRV)

		changePageCmd = tea.Batch(tea.ClearScreen, changePageCmd)
	}

	if m.activePageIdx != int(Page_ContactUs) {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+j":
				changePageCmd = tea.Batch(changePageCmd, m.ChangePage(Page_Jobs))
			case "esc":
				if !m.ignoreESC && len(m.history) > 0 {
					var page PageEnum
					l := len(m.history)
					m.history, page = m.history[:l-1], m.history[l-1]

					m.pages[m.activePageIdx] = m.pages[m.activePageIdx].OnClose()
					m.activePageIdx = int(page)
					m.pages[m.activePageIdx] = m.pages[m.activePageIdx].OnOpen()
					wsMsg := tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					}
					model, updateSizeCmd := m.pages[m.activePageIdx].Update(wsMsg)
					m.pages[m.activePageIdx] = model.(Page)
					changePageCmd = tea.Batch(changePageCmd, updateSizeCmd)

					newRV := m.pages[m.activePageIdx].SetSize(wsMsg)
					m.pages[m.activePageIdx] = m.pages[m.activePageIdx].SetResponsiveView(newRV)
					changePageCmd = tea.Batch(tea.ClearScreen, changePageCmd)

				}
			}
		}
	}

	currentPageIdx := m.activePageIdx // it might change during update
	model, cmd := m.pages[currentPageIdx].Update(msg)
	if changePageCmd != nil {
		cmd = tea.Batch(cmd, changePageCmd)
	}
	m.pages[currentPageIdx] = model.(Page)

	return m, tea.Batch(cmd, TickCmdWithDuration(200*time.Microsecond))
}

func (m *App) View() string {
	return m.pages[m.activePageIdx].View()
}

func (m *App) SetIgnoreEsc(b bool) {
	m.ignoreESC = b
}
