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
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

type Row []string
type Rows []Row

func (r Row) ToTableRow() table.Row {
	data := table.RowData{}
	for idx, d := range r {
		data[fmt.Sprintf("%d", idx)] = d
	}
	return table.NewRow(data)
}

func (r Rows) ToTableRows() []table.Row {
	var rows []table.Row
	for _, row := range r {
		rows = append(rows, row.ToTableRow())
	}
	return rows
}

type ResourceDetailsPage struct {
	item                *golang.OptimizationItem
	deviceTable         table.Model
	detailTable         table.Model
	deviceProperties    map[string]Rows
	selectedDevice      string
	detailTableHasFocus bool

	helpController          *controller.Help
	optimizationsController *controller.Optimizations[golang.OptimizationItem]
	statusBar               StatusBarView
	app                     *App
	responsive.ResponsiveView
	detailColumns []table.Column
}

func (m ResourceDetailsPage) ExtractProperties(item *golang.OptimizationItem) map[string]Rows {
	res := map[string]Rows{}
	for _, dev := range item.Devices {
		rows := Rows{}

		for _, prop := range dev.Properties {
			if !strings.HasPrefix(prop.Key, " ") {
				prop.Key = style.Bold.Render(prop.Key)
			}
			rows = append(rows, Row{
				prop.Key,
				prop.Current,
				prop.Average,
				prop.Recommended,
			})
		}
		res[dev.DeviceId] = rows
	}

	for deviceID, rows := range res {
		for idx, row := range rows {
			if row[1] != row[3] {
				row[1] = style.ChangeFrom.Render(row[1])
				row[3] = style.ChangeTo.Render(row[3])
			}
			rows[idx] = row
		}
		res[deviceID] = rows
	}
	return res
}

func NewOptimizationDetailsView(
	optimizationsController *controller.Optimizations[golang.OptimizationItem],
	helpController *controller.Help,
	statusBar StatusBarView,
) ResourceDetailsPage {
	return ResourceDetailsPage{
		helpController:          helpController,
		optimizationsController: optimizationsController,
		statusBar:               statusBar,
	}
}

func (m ResourceDetailsPage) OnOpen() Page {
	item := m.optimizationsController.SelectedItem()

	ifRecommendationExists := func(f func() string) string {
		if !item.Loading && !item.Skipped && !item.LazyLoadingEnabled {
			return f()
		}
		return ""
	}

	deviceColumns := []table.Column{
		table.NewColumn("0", "Resource ID", 0),
		table.NewColumn("1", "Resource Name", 0),
		table.NewColumn("2", "ResourceType", 0),
		table.NewColumn("3", "Runtime", 0),
		table.NewColumn("4", "Current Cost", 0),
		table.NewColumn("5", "Right sized Cost", 0),
		table.NewColumn("6", "Savings", 0),
	}

	deviceRows := Rows{}
	for _, dev := range item.Devices {
		deviceRows = append(deviceRows, Row{
			dev.DeviceId,
			item.Name,
			dev.ResourceType,
			dev.Runtime,
			fmt.Sprintf("%s", utils.FormatPriceFloat(dev.CurrentCost)),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("%s", utils.FormatPriceFloat(dev.RightSizedCost))
			}),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("%s", utils.FormatPriceFloat(dev.CurrentCost-dev.RightSizedCost))
			}),
		})
	}

	for idx, column := range deviceColumns {
		width := len(column.Title())
		for _, row := range deviceRows.ToTableRows() {
			cell := row.Data[column.Key()]
			cellContent := ""
			if cell != nil {
				cellContent = strings.TrimSpace(style.StyleSelector.ReplaceAllString(cell.(string), ""))
			}
			if len(cellContent) > width {
				width = len(cellContent)
			}
		}
		column = table.NewColumn(column.Key(), column.Title(), width)
		deviceColumns[idx] = column
	}

	days := "7"
	for _, p := range item.Preferences {
		if p.Key == "ObservabilityTimePeriod" && p.Value != nil {
			days = p.Value.Value
		}
	}
	m.detailColumns = []table.Column{
		table.NewColumn("0", "", 30),
		table.NewColumn("1", "Current", 30),
		table.NewColumn("2", fmt.Sprintf("%s day usage", days), 15),
		table.NewColumn("3", "Recommendation", 30),
	}

	m.item = item
	m.detailTable = table.New(m.detailColumns).
		WithPageSize(1).
		WithHorizontalFreezeColumnCount(1).
		WithBaseStyle(style.Base).BorderRounded()
	m.deviceTable = table.New(deviceColumns).
		WithRows(deviceRows.ToTableRows()).
		WithHighlightedRow(0).
		WithHorizontalFreezeColumnCount(1).
		Focused(true).
		WithPageSize(len(deviceRows)).
		WithBaseStyle(style.ActiveStyleBase).BorderRounded()
	m.deviceProperties = m.ExtractProperties(item)
	m.detailTableHasFocus = false
	m.selectedDevice = ""
	m.helpController.SetKeyMap([]string{
		"↑/↓: move",
		"←/→: scroll in the table",
		"enter: switch to device detail table",
		"esc: back to optimizations list",
		"q/ctrl+c: exit",
	})
	return m
}
func (m ResourceDetailsPage) OnClose() Page {
	return m
}
func (m ResourceDetailsPage) Init() tea.Cmd {
	return nil
}

func (m ResourceDetailsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, detailCMD tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.detailTable = m.detailTable.WithMaxTotalWidth(m.GetWidth())
		m.deviceTable = m.deviceTable.WithMaxTotalWidth(m.GetWidth())
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			m.detailTableHasFocus = true
			m.app.SetIgnoreEsc(true)
			m.deviceTable = m.deviceTable.WithBaseStyle(style.Base)
			m.detailTable = m.detailTable.WithBaseStyle(style.ActiveStyleBase).Focused(true).WithHighlightedRow(0)
		case "right":
			if m.detailTableHasFocus {
				m.detailTable = m.detailTable.ScrollRight()
			} else {
				m.deviceTable = m.deviceTable.ScrollRight()
			}
		case "left":
			if m.detailTableHasFocus {
				m.detailTable = m.detailTable.ScrollLeft()
			} else {
				m.deviceTable = m.deviceTable.ScrollLeft()
			}
		case "esc":
			m.app.SetIgnoreEsc(false)
			if m.detailTableHasFocus {
				m.detailTableHasFocus = false
				m.detailTable = m.detailTable.Focused(false).WithBaseStyle(style.Base)
				m.deviceTable = m.deviceTable.Focused(true).WithBaseStyle(style.ActiveStyleBase)
			}
		default:
			if m.detailTableHasFocus {
				m.detailTable, cmd = m.detailTable.Update(msg)
			} else {
				m.deviceTable, cmd = m.deviceTable.Update(msg)
			}
		}
	default:
		if m.detailTableHasFocus {
			m.detailTable, cmd = m.detailTable.Update(msg)
		} else {
			m.deviceTable, cmd = m.deviceTable.Update(msg)
		}
	}

	if m.deviceTable.HighlightedRow().Data["0"] != nil && m.selectedDevice != m.deviceTable.HighlightedRow().Data["0"] {
		m.selectedDevice = m.deviceTable.HighlightedRow().Data["0"].(string)
		rows := m.deviceProperties[m.selectedDevice].ToTableRows()
		var columns []table.Column
		for _, column := range m.detailColumns {
			width := len(column.Title())
			for _, row := range rows {
				cell := row.Data[column.Key()]
				cellContent := ""
				if cell != nil {
					cellContent = strings.TrimSpace(style.StyleSelector.ReplaceAllString(cell.(string), ""))
				}
				if len(cellContent) > width {
					width = len(cellContent)
				}
			}
			columns = append(columns, table.NewColumn(column.Key(), column.Title(), width+2))
		}
		m.detailTable = m.detailTable.WithColumns(columns).WithRows(rows)
	}

	lineCount := strings.Count(wordwrap.String(m.item.Description, m.GetWidth()), "\n") + 1
	deviceTableHeight := 7
	detailsTableHeight := 7

	for lineCount+deviceTableHeight+detailsTableHeight+m.statusBar.Height() < m.GetHeight() {
		if deviceTableHeight-6 < len(m.deviceProperties) && deviceTableHeight < detailsTableHeight {
			deviceTableHeight++
		} else {
			detailsTableHeight++
		}
	}
	m.deviceTable = m.deviceTable.WithPageSize(deviceTableHeight - 6).WithMaxTotalWidth(m.GetWidth())
	m.detailTable = m.detailTable.WithPageSize(detailsTableHeight - 6).WithMaxTotalWidth(m.GetWidth())
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)

	return m, tea.Batch(detailCMD, cmd)
}

func (m ResourceDetailsPage) View() string {
	return m.deviceTable.View() + "\n" +
		wordwrap.String(m.item.Description, m.GetWidth()) + "\n" +
		m.detailTable.View() + "\n" +
		m.statusBar.View()
}

func (m ResourceDetailsPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
func (m ResourceDetailsPage) SetApp(app *App) ResourceDetailsPage {
	m.app = app
	return m
}
