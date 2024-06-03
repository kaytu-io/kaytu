package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/view/responsive"
)

type OverviewPage struct {
	table       table.Model
	clearScreen bool

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
		table.NewColumn("0", "Resource Id", 23),
		table.NewColumn("1", "Resource Name", 23),
		table.NewColumn("2", "Resource Type", 15),
		table.NewColumn("3", "Region", 15),
		table.NewColumn("4", "Platform", 15),
		table.NewColumn("5", "Total Saving (Monthly)", 40),
		table.NewColumn("6", "", 1),
	}

	t := table.New(columns).
		Focused(true).
		WithPageSize(10).
		WithHorizontalFreezeColumnCount(1).
		WithBaseStyle(style.ActiveStyleBase).
		BorderRounded().
		HighlightStyle(style.HighlightStyle)

	return OverviewPage{
		optimizations:  optimizations,
		helpController: helpController,
		table:          t,
		statusBar:      statusBar,
	}
}

func (m OverviewPage) OnClose() Page {
	return m
}

func (m OverviewPage) OnOpen() Page {
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

func (m OverviewPage) Init() tea.Cmd {
	return nil
}

func (m OverviewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case "right":
			m.table = m.table.ScrollRight()
		case "left":
			m.table = m.table.ScrollLeft()
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

	return fmt.Sprintf("Current runtime cost: %s, Savings: %s\n%s\n%s",
		style.CostStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(totalCost))), style.SavingStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(savings))),
		m.table.View(),
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
