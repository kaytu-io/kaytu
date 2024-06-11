package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view/responsive"
	"math"
	"sort"
	"strings"
)

type PluginCustomOverviewPage struct {
	table       table.Model
	clearScreen bool

	filterInput   textinput.Model
	focusOnFilter bool
	sortColumnIdx int
	sortDesc      bool
	rows          []table.Row

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
		columns = append(columns, table.NewColumn(column.GetId(), column.GetName(), int(column.GetWidth())).WithFiltered(true))
		tableColumnIdToIndex[column.GetId()] = i
	}
	filterInput := textinput.New()
	t := table.New(columns).
		Focused(true).
		WithPageSize(10).
		WithHorizontalFreezeColumnCount(1).
		WithBaseStyle(style.ActiveStyleBase).
		BorderRounded().
		Filtered(true).
		HighlightStyle(style.HighlightStyle)

	return PluginCustomOverviewPage{
		filterInput:          filterInput,
		optimizations:        optimizations,
		helpController:       helpController,
		table:                t,
		statusBar:            statusBar,
		chartDefinition:      chartDefinition,
		chartDefinitionDirty: false,
		sortColumnIdx:        -1,
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
		"pgdown/pgup/shift+↑/↓: next/prev page",
		"home/end/shift+h/shift+e: first/last page",
		"←/→: scroll in the table",
		"enter: see resource details",
		"p: change preferences",
		"P: change preferences for all resources",
		"r: load all items in current page",
		"shift+r: load all items",
		"/: filter results",
		"s: sort by next column",
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
	dontSendUpdateToTable := false

	if m.focusOnFilter {
		var filterCmd tea.Cmd
		m.filterInput, filterCmd = m.filterInput.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "esc", "enter":
				m.app.SetIgnoreEsc(false)
				m.filterInput.Blur()
				m.focusOnFilter = false
			}
		}
		m.table = m.table.WithFilterInputValue(m.filterInput.Value())
		return m, filterCmd
	}

	var rows RowsWithId

	if m.chartDefinitionDirty {
		var columns []table.Column
		for idx, column := range m.chartDefinition.GetColumns() {
			name := column.GetName()
			if idx == m.sortColumnIdx {
				if m.sortDesc {
					name = name + " ↓"
				} else {
					name = name + " ↑"
				}
			}
			width := len(name)
			for _, row := range m.rows {
				cell := row.Data[column.GetId()]
				cellContent := ""
				if cell != nil {
					cellContent = strings.TrimSpace(style.StyleSelector.ReplaceAllString(cell.(string), ""))
				}
				if len(cellContent) > width {
					width = len(cellContent)
				}
			}
			columns = append(columns, table.NewColumn(column.GetId(), name, width).WithFiltered(true))
		}
		m.table = m.table.WithColumns(columns)
		m.chartDefinitionDirty = false
	}

	sortColumn := ""
	if m.sortColumnIdx > 0 {
		sortColumn = m.chartDefinition.Columns[m.sortColumnIdx].Id
	}
	for _, i := range m.optimizations.Items() {
		rowValues := make(map[string]string)
		sortValue := math.MaxFloat64
		for k, value := range i.GetOverviewChartRow().GetValues() {
			if sortColumn == k {
				sortValue = value.GetSortValue()
			}
			for _, column := range m.chartDefinition.GetColumns() {
				if column.GetId() == k {
					if column.Width < uint32(len(value.GetValue())) {
						m.chartDefinitionDirty = true
					}
				}
			}
			rowValues[k] = strings.TrimSpace(value.GetValue())
		}
		row := RowWithId{
			ID:        i.GetOverviewChartRow().GetRowId(),
			SortValue: sortValue,
			Row:       rowValues,
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].SortValue == math.MaxFloat64 {
			return false
		}
		if rows[j].SortValue == math.MaxFloat64 {
			return true
		}

		if m.sortDesc {
			return rows[i].SortValue > rows[j].SortValue
		} else {
			return rows[i].SortValue < rows[j].SortValue
		}
	})
	m.rows = rows.ToTableRows()
	m.table = m.table.WithRows(m.rows)
	m.table = m.table.WithFilterInputValue(m.filterInput.Value())

	var changePageCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "shift+down":
			m.table = m.table.PageDown()
		case "shift+up":
			m.table = m.table.PageUp()
		case "home", "H":
			m.table = m.table.PageFirst()
		case "end", "E":
			m.table = m.table.PageLast()
		case "s":
			if m.sortColumnIdx != -1 && !m.sortDesc {
				m.sortDesc = true
			} else {
				newSortIdx := m.sortColumnIdx
				for {
					newSortIdx++
					if newSortIdx >= len(m.chartDefinition.Columns) {
						newSortIdx = -1
						break
					}

					if m.chartDefinition.Columns[newSortIdx].Sortable {
						break
					}

					if newSortIdx == m.sortColumnIdx {
						break
					}
				}
				m.sortColumnIdx = newSortIdx
				m.sortDesc = false
			}
			m.chartDefinitionDirty = true

		case "/":
			m.focusOnFilter = true
			m.filterInput.Focus()
			m.app.SetIgnoreEsc(true)
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
			dontSendUpdateToTable = true
		case "left":
			m.table = m.table.ScrollLeft()
			dontSendUpdateToTable = true
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
	if !dontSendUpdateToTable {
		m.table, cmd = m.table.Update(msg)
	}

	if changePageCmd != nil {
		cmd = tea.Batch(cmd, changePageCmd)
	}
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)
	m.statusBar.initialization = m.optimizations.GetInitialization()

	m.table = m.table.WithPageSize(m.GetHeight() - (8 + m.statusBar.Height())).WithMaxTotalWidth(m.GetWidth())
	return m, cmd
}

func (m *PluginCustomOverviewPage) View() string {
	return fmt.Sprintf("%s\n%s\nFilter: %s\n%s",
		m.optimizations.GetResultSummary(),
		m.table.View(),
		m.filterInput.View(),
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
