package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view/responsive"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

const XKaytuRowId = "x_kaytu_row_id"

type RowsWithId []RowWithId

func (r RowsWithId) ToTableRows() []table.Row {
	var rows []table.Row
	for _, row := range r {
		rows = append(rows, row.ToTableRow())
	}
	return rows
}

type RowWithId struct {
	ID  string
	Row map[string]string
}

func (r RowWithId) ToTableRow() table.Row {
	data := table.RowData{}
	for k, v := range r.Row {
		data[k] = v
	}
	data[XKaytuRowId] = r.ID
	return table.NewRow(data)
}

type PluginCustomResourceDetailsPage struct {
	chartDefinition      *golang.ChartDefinition
	chartDefinitionDirty bool

	item                *golang.ChartOptimizationItem
	deviceTable         table.Model
	detailTable         table.Model
	deviceProperties    map[string]Rows
	selectedDevice      string
	detailTableHasFocus bool

	helpController          *controller.Help
	optimizationsController *controller.Optimizations[golang.ChartOptimizationItem]
	statusBar               StatusBarView
	app                     *App
	responsive.ResponsiveView
}

func (m *PluginCustomResourceDetailsPage) ExtractProperties(item *golang.ChartOptimizationItem) map[string]Rows {
	res := map[string]Rows{}
	for devId, dev := range item.GetDevicesProperties() {
		rows := Rows{}

		for _, prop := range dev.Properties {
			if !strings.HasPrefix(prop.Key, " ") {
				prop.Key = style.Bold.Render(prop.Key)
			}
			usageColumn := ""
			if prop.Average != "" {
				usageColumn += fmt.Sprintf("Avg: %s", prop.Average)
			}
			if prop.Max != "" {
				if usageColumn != "" {
					usageColumn += " | "
				}
				usageColumn += fmt.Sprintf("Max: %s", prop.Max)
			}
			rows = append(rows, Row{
				prop.Key,
				prop.Current,
				usageColumn,
				prop.Recommended,
			})
		}
		res[devId] = rows
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

func NewPluginCustomOptimizationDetailsView(
	chartDefinition *golang.ChartDefinition,
	optimizationsController *controller.Optimizations[golang.ChartOptimizationItem],
	helpController *controller.Help,
	statusBar StatusBarView,
) PluginCustomResourceDetailsPage {
	return PluginCustomResourceDetailsPage{
		chartDefinition:         chartDefinition,
		helpController:          helpController,
		optimizationsController: optimizationsController,
		statusBar:               statusBar,
	}
}

func (m *PluginCustomResourceDetailsPage) SetChartDefinition(chartDefinition *golang.ChartDefinition) {
	m.chartDefinition = chartDefinition
	m.chartDefinitionDirty = true
}

func (m *PluginCustomResourceDetailsPage) OnOpen() Page {
	item := m.optimizationsController.SelectedItem()

	var deviceColumns []table.Column
	for _, column := range m.chartDefinition.GetColumns() {
		deviceColumns = append(deviceColumns, table.NewColumn(column.GetId(), column.GetName(), int(column.GetWidth())))
	}

	deviceRows := RowsWithId{}
	for _, deviceChartRow := range item.GetDevicesChartRows() {
		rowValues := make(map[string]string)
		for key, value := range deviceChartRow.GetValues() {
			rowValues[key] = value.GetValue()
		}

		deviceRows = append(deviceRows, RowWithId{
			ID:  deviceChartRow.GetRowId(),
			Row: rowValues,
		})
	}

	days := "7"
	for _, p := range item.Preferences {
		if p.Key == "ObservabilityTimePeriod" && p.Value != nil {
			days = p.Value.Value
		}
	}
	detailColumns := []table.Column{
		table.NewColumn("0", "", 30),
		table.NewColumn("1", "Current", 30),
		table.NewColumn("2", fmt.Sprintf("%s day usage", days), 15),
		table.NewColumn("3", "Recommendation", 30),
	}

	m.item = item
	m.detailTable = table.New(detailColumns).
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
	m.chartDefinitionDirty = false
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
func (m *PluginCustomResourceDetailsPage) OnClose() Page {
	return m
}
func (m *PluginCustomResourceDetailsPage) Init() tea.Cmd {
	return nil
}

func (m *PluginCustomResourceDetailsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.chartDefinitionDirty {
		var columns []table.Column
		for _, column := range m.chartDefinition.GetColumns() {
			columns = append(columns, table.NewColumn(column.GetId(), column.GetName(), int(column.GetWidth())))
		}
		m.deviceTable = m.deviceTable.WithColumns(columns)
		m.chartDefinitionDirty = false
	}

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

	if m.deviceTable.HighlightedRow().Data[XKaytuRowId] != nil && m.selectedDevice != m.deviceTable.HighlightedRow().Data[XKaytuRowId] {
		m.selectedDevice = m.deviceTable.HighlightedRow().Data[XKaytuRowId].(string)

		m.detailTable = m.detailTable.WithRows(m.deviceProperties[m.selectedDevice].ToTableRows())
	}

	lineCount := strings.Count(wordwrap.String(m.item.GetDescription(), m.GetWidth()), "\n") + 1
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

func (m *PluginCustomResourceDetailsPage) View() string {
	return m.deviceTable.View() + "\n" +
		wordwrap.String(m.item.GetDescription(), m.GetWidth()) + "\n" +
		m.detailTable.View() + "\n" +
		m.statusBar.View()
}

func (m *PluginCustomResourceDetailsPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
func (m *PluginCustomResourceDetailsPage) SetApp(app *App) *PluginCustomResourceDetailsPage {
	m.app = app
	return m
}
