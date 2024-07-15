package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/view/responsive"
	"strings"
)

type OverviewPage struct {
	table       table.Model
	clearScreen bool

	filterInput   textinput.Model
	focusOnFilter bool
	sortColumnIdx int
	sortDesc      bool
	columns       []table.Column

	helpController *controller.Help
	optimizations  *controller.Optimizations[golang.OptimizationItem]
	statusBar      StatusBarView
	app            *App

	responsive.ResponsiveView
}

func NewOptimizationsView(
	optimizations *controller.Optimizations[golang.OptimizationItem],
	helpController *controller.Help,
	statusBar StatusBarView,
) OverviewPage {
	columns := []table.Column{
		table.NewColumn("0", "Resource Id", 23).WithFiltered(true),
		table.NewColumn("1", "Resource Name", 23).WithFiltered(true),
		table.NewColumn("2", "Resource Type", 15).WithFiltered(true),
		table.NewColumn("3", "Region", 15).WithFiltered(true),
		table.NewColumn("4", "Platform", 15).WithFiltered(true),
		table.NewColumn("5", "Total Saving (Monthly)", 40).WithFiltered(true),
		table.NewColumn("6", "", 1),
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

	return OverviewPage{
		filterInput:    filterInput,
		optimizations:  optimizations,
		helpController: helpController,
		table:          t,
		columns:        columns,
		statusBar:      statusBar,
	}
}

func (m OverviewPage) OnClose() Page {
	return m
}

func (m OverviewPage) OnOpen() Page {
	m.helpController.SetKeyMap([]string{
		"↑/↓: move",
		"←/→: scroll in the table",
		"enter: see resource details",
		"p: change preferences",
		"P: change preferences for all resources",
		"r: load all items in current page",
		"shift+r: load all items",
		"s: change sort",
		"q/ctrl+c: exit",
	})
	return m
}

func (m OverviewPage) Init() tea.Cmd {
	return nil
}

func (m OverviewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dontSendUpdateToTable := false
	var filterCmd tea.Cmd
	if m.focusOnFilter {
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

	var rows Rows
	for _, i := range m.optimizations.Items() {
		totalSaving := 0.0
		totalCurrentCost := 0.0
		if !i.Loading && !i.Skipped && !i.LazyLoadingEnabled {
			for _, dev := range i.Devices {
				totalSaving += dev.CurrentCost - dev.RightSizedCost
				totalCurrentCost += dev.CurrentCost
			}
		}

		row := Row{
			i.Id,
			i.Name,
			i.ResourceType,
			i.Region,
			i.Platform,
			fmt.Sprintf("%s (%.2f%%)", utils.FormatPriceFloat(totalSaving), (totalSaving/totalCurrentCost)*100),
		}
		if i.Skipped {
			row[5] = "skipped"
			if len(i.SkipReason) > 0 {
				row[5] += " - " + i.SkipReason
			}
		} else if i.LazyLoadingEnabled {
			row[5] = "press enter to load"
		} else if i.Loading {
			row[5] = "loading"
		}
		row = append(row, "→")
		rows = append(rows, row)
	}
	var columns []table.Column
	for idx, column := range m.columns {
		width := len(column.Title())
		for _, row := range rows.ToTableRows() {
			cell := row.Data[column.Key()]
			cellContent := ""
			if cell != nil {
				cellContent = strings.TrimSpace(style.StyleSelector.ReplaceAllString(cell.(string), ""))
			}
			if len(cellContent) > width {
				width = len(cellContent)
			}
		}
		if idx == 6 {
			width = -1
		}
		columns = append(columns, table.NewColumn(column.Key(), column.Title(), width+2).WithFiltered(true))
	}

	m.table = m.table.WithColumns(columns).WithRows(rows.ToTableRows())
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
		case "home", "shift+h":
			m.table = m.table.PageFirst()
		case "end", "shift+e":
			m.table = m.table.PageLast()
		case "q":
			return m, tea.Quit
		case "p":
			if m.table.TotalRows() == 0 {
				break
			}
			selectedInstanceID := m.table.HighlightedRow().Data["0"]
			for _, i := range m.optimizations.Items() {
				if selectedInstanceID == i.Id && !i.Skipped && !i.Loading && !i.LazyLoadingEnabled {
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
				if !i.Skipped && i.LazyLoadingEnabled {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.Id, i.Preferences)
				}
			}

		case "R":
			for _, i := range m.optimizations.Items() {
				if !i.Skipped && i.LazyLoadingEnabled {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.Id, i.Preferences)
				}
			}
		case "s":
			if m.sortDesc {
				m.sortDesc = false
				m.table = m.table.SortByAsc(fmt.Sprintf("%d", m.sortColumnIdx))
			} else {
				m.sortColumnIdx = (m.sortColumnIdx + 1) % 6
				m.sortDesc = true
				m.table = m.table.SortByDesc(fmt.Sprintf("%d", m.sortColumnIdx))
			}

			var columns []table.Column
			for idx, col := range m.columns {
				name := col.Title()
				if m.sortColumnIdx == idx {
					if m.sortDesc {
						name = name + " ↓"
					} else {
						name = name + " ↑"
					}
				}
				columns = append(columns, table.NewColumn(col.Key(), name, col.Width()).WithFiltered(true))
			}
			m.table = m.table.WithColumns(columns)

		case "/":
			m.focusOnFilter = true
			m.filterInput.Focus()
			m.app.SetIgnoreEsc(true)

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

			selectedInstanceID := m.table.HighlightedRow().Data["0"]
			for _, i := range m.optimizations.Items() {
				if selectedInstanceID == i.Id && !i.Skipped && !i.Loading && !i.LazyLoadingEnabled {
					m.optimizations.SelectItem(i)
					changePageCmd = m.app.ChangePage(Page_ResourceDetails)
					break
				} else if selectedInstanceID == i.Id && !i.Skipped && i.LazyLoadingEnabled {
					i.LazyLoadingEnabled = false
					i.Loading = true
					m.optimizations.SendItem(i)
					m.optimizations.ReEvaluate(i.Id, i.Preferences)
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

func (m OverviewPage) View() string {
	//if m.clearScreen {
	//	m.clearScreen = false
	//	return ""
	//}

	totalCost := 0.0
	savings := 0.0
	for _, i := range m.optimizations.Items() {
		for _, dev := range i.Devices {
			totalCost += dev.CurrentCost
			savings += dev.CurrentCost - dev.RightSizedCost
		}
	}

	return fmt.Sprintf("Current runtime cost: %s, Savings: %s\n%s\nFilter: %s\n%s",
		style.CostStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(totalCost))), style.SavingStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(savings))),
		m.table.View(),
		m.filterInput.View(),
		m.statusBar.View(),
	)
}

func (m OverviewPage) SetApp(app *App) OverviewPage {
	m.app = app
	return m
}

func (m OverviewPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
