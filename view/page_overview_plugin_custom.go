package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view/responsive"
)

type PluginCustomOverviewPage struct {
	table       table.Model
	clearScreen bool

	helpController *controller.Help
	optimizations  *controller.Optimizations[golang.ChartOptimizationItem]
	statusBar      StatusBarView
	app            *App

	chartDefinition      *golang.ChartDefinition
	chartDefinitionDirty bool

	responsive.ResponsiveView
}

func NewPluginCustomOverviewPageView(
	chartDefinition *golang.ChartDefinition,
	optimizations *controller.Optimizations[golang.ChartOptimizationItem],
	helpController *controller.Help,
	statusBar StatusBarView,
) PluginCustomOverviewPage {
	var columns []table.Column
	tableColumnIdToIndex := make(map[string]int)
	for i, column := range chartDefinition.GetColumns() {
		columns = append(columns, table.NewColumn(column.GetId(), column.GetName(), int(column.GetWidth())))
		tableColumnIdToIndex[column.GetId()] = i
	}
	t := table.New(columns).
		Focused(true).
		WithPageSize(10).
		WithHorizontalFreezeColumnCount(1).
		WithBaseStyle(style.ActiveStyleBase).
		BorderRounded().
		HighlightStyle(style.HighlightStyle)

	return PluginCustomOverviewPage{
		optimizations:        optimizations,
		helpController:       helpController,
		table:                t,
		statusBar:            statusBar,
		chartDefinition:      chartDefinition,
		chartDefinitionDirty: false,
	}
}

func (m *PluginCustomOverviewPage) SetChartDefinition(chartDefinition *golang.ChartDefinition) {
	m.chartDefinition = chartDefinition
	m.chartDefinitionDirty = true
}

func (m *PluginCustomOverviewPage) OnClose() Page {
	return m
}

func (m *PluginCustomOverviewPage) OnOpen() Page {
	m.helpController.SetKeyMap([]string{
		"↑/↓: move",
		"pgdown/pgup: next/prev page",
		"←/→: scroll in the table",
		"enter: see resource details",
		"p: change preferences",
		"P: change preferences for all resources",
		"r: load all items in current page",
		"shift+r: load all items in all pages",
		"ctrl+j: list of jobs",
		"q/ctrl+c: exit",
	})
	return m
}

func (m *PluginCustomOverviewPage) Init() tea.Cmd {
	return nil
}

func (m *PluginCustomOverviewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var rows RowsWithId

	if m.chartDefinitionDirty {
		var columns []table.Column
		for _, column := range m.chartDefinition.GetColumns() {
			columns = append(columns, table.NewColumn(column.GetId(), column.GetName(), int(column.GetWidth())))
		}
		m.table = m.table.WithColumns(columns)
		m.chartDefinitionDirty = false
	}

	for _, i := range m.optimizations.Items() {
		rowValues := make(map[string]string)
		for k, value := range i.GetOverviewChartRow().GetValues() {
			rowValues[k] = value.GetValue()
		}
		row := RowWithId{
			ID:  i.GetOverviewChartRow().GetRowId(),
			Row: rowValues,
		}
		rows = append(rows, row)
	}
	m.table = m.table.WithRows(rows.ToTableRows())

	var changePageCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "pgdown":
			m.table.PageDown()
		case "pgup":
			m.table.PageUp()
		case "home":
			m.table.PageFirst()
		case "end":
			m.table.PageLast()
		case "q":
			return m, tea.Quit
		case "p":
			if m.table.TotalRows() == 0 {
				break
			}
			selectedRowId := m.table.HighlightedRow().Data[XKaytuRowId]
			for _, i := range m.optimizations.Items() {
				if selectedRowId == i.GetOverviewChartRow().GetRowId() && !i.GetSkipped() && !i.GetLoading() && !i.GetLazyLoadingEnabled() {
					m.optimizations.SelectItem(i)
					changePageCmd = m.app.ChangePage(Page_Preferences)
					m.clearScreen = true
					break
				}
			}

		case "P":
			if m.table.TotalRows() == 0 {
				break
			}
			m.optimizations.SelectItem(nil)
			changePageCmd = m.app.ChangePage(Page_Preferences)
			m.clearScreen = true

		case "r":
			start, end := m.table.VisibleIndices()
			for _, i := range m.optimizations.Items()[start : end+1] {
				if !i.GetSkipped() && i.GetLazyLoadingEnabled() {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.GetOverviewChartRow().GetRowId(), i.GetPreferences())
				}
			}

		case "R":
			for _, i := range m.optimizations.Items() {
				if !i.GetSkipped() && i.GetLazyLoadingEnabled() {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.GetOverviewChartRow().GetRowId(), i.GetPreferences())
				}
			}

		case "right":
			m.table = m.table.ScrollRight()
		case "left":
			m.table = m.table.ScrollLeft()
		case "enter":
			if m.table.TotalRows() == 0 {
				break
			}

			selectedRowId := m.table.HighlightedRow().Data[XKaytuRowId].(string)
			for _, i := range m.optimizations.Items() {
				if selectedRowId == i.GetOverviewChartRow().GetRowId() && !i.GetSkipped() && !i.GetLoading() && !i.GetLazyLoadingEnabled() {
					m.optimizations.SelectItem(i)
					changePageCmd = m.app.ChangePage(Page_ResourceDetails)
					break
				} else if selectedRowId == i.GetOverviewChartRow().GetRowId() && !i.GetLoading() && i.GetLazyLoadingEnabled() {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.GetOverviewChartRow().GetRowId(), i.GetPreferences())
				}
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	if changePageCmd != nil {
		cmd = tea.Batch(cmd, changePageCmd)
	}
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)
	m.statusBar.initialization = m.optimizations.GetInitialization()

	m.table = m.table.WithPageSize(m.GetHeight() - (7 + m.statusBar.Height())).WithMaxTotalWidth(m.GetWidth())

	return m, cmd
}

func (m *PluginCustomOverviewPage) View() string {
	return fmt.Sprintf("%s\n%s\n%s",
		m.optimizations.GetResultSummary(),
		m.table.View(),
		m.statusBar.View(),
	)
}

func (m *PluginCustomOverviewPage) SetApp(app *App) *PluginCustomOverviewPage {
	m.app = app
	return m
}

func (m *PluginCustomOverviewPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
