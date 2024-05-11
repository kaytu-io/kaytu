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

type OptimizationDetailsPage struct {
	item                *golang.OptimizationItem
	deviceTable         table.Model
	detailTable         table.Model
	deviceProperties    map[string]Rows
	selectedDevice      string
	detailTableHasFocus bool

	helpController          *controller.Help
	optimizationsController *controller.Optimizations
	statusBar               StatusBarView
	responsive.ResponsiveView
}

func ExtractProperties(item *golang.OptimizationItem) map[string]Rows {
	res := map[string]Rows{}
	for _, dev := range item.Devices {
		rows := Rows{
			{
				"",
				"",
				style.Bold.Render("Average"),
				style.Bold.Render("Max"),
				"",
			},
		}

		for _, prop := range dev.Properties {
			if !strings.HasPrefix(prop.Key, " ") {
				prop.Key = style.Bold.Render(prop.Key)
			}
			rows = append(rows, Row{
				prop.Key,
				prop.Current,
				prop.Average,
				prop.Max,
				prop.Recommended,
			})
		}
		res[dev.DeviceId] = rows
	}

	for deviceID, rows := range res {
		for idx, row := range rows {
			if row[1] != row[4] {
				row[1] = style.ChangeFrom.Render(row[1])
				row[4] = style.ChangeTo.Render(row[4])
			}
			rows[idx] = row
		}
		res[deviceID] = rows
	}
	return res
}

func NewOptimizationDetailsView(
	optimizationsController *controller.Optimizations,
	helpController *controller.Help,
	statusBar StatusBarView,
) OptimizationDetailsPage {
	return OptimizationDetailsPage{
		helpController:          helpController,
		optimizationsController: optimizationsController,
		statusBar:               statusBar,
	}
}

func (m OptimizationDetailsPage) OnOpen() Page {
	item := m.optimizationsController.SelectedItem()

	ifRecommendationExists := func(f func() string) string {
		if !item.Loading && !item.Skipped && !item.LazyLoadingEnabled {
			return f()
		}
		return ""
	}

	deviceColumns := []table.Column{
		table.NewColumn("0", "DeviceID", 30),
		table.NewColumn("1", "ResourceType", 20),
		table.NewColumn("2", "Runtime", 13),
		table.NewColumn("3", "Current Cost", 20),
		table.NewColumn("4", "Right sized Cost", 20),
		table.NewColumn("5", "Savings", 20),
	}

	deviceRows := Rows{}
	for _, dev := range item.Devices {
		deviceRows = append(deviceRows, Row{
			dev.DeviceId,
			dev.ResourceType,
			dev.Runtime,
			fmt.Sprintf("$%.2f", dev.CurrentCost),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("$%.2f", dev.RightSizedCost)
			}),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("$%.2f", dev.CurrentCost-dev.RightSizedCost)
			}),
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
		table.NewColumn("3", "", 15),
		table.NewColumn("4", "Recommendation", 30),
	}

	m.item = item
	m.detailTable = table.New(detailColumns).
		WithPageSize(1).
		WithBaseStyle(style.Base).BorderRounded()
	m.deviceTable = table.New(deviceColumns).
		WithRows(deviceRows.ToTableRows()).
		WithHighlightedRow(0).
		Focused(true).
		WithPageSize(len(deviceRows)).
		WithBaseStyle(style.ActiveStyleBase).BorderRounded()
	m.deviceProperties = ExtractProperties(item)
	m.detailTableHasFocus = false
	m.selectedDevice = ""
	m.helpController.SetKeyMap([]string{
		"↑/↓: move",
		"esc/←: back to optimizations list",
		"q/ctrl+c: exit",
	})
	return m
}
func (m OptimizationDetailsPage) OnClose() Page {
	return m
}
func (m OptimizationDetailsPage) Init() tea.Cmd {
	return nil
}

func (m OptimizationDetailsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.deviceTable = m.deviceTable.WithBaseStyle(style.Base)
			m.detailTable = m.detailTable.WithBaseStyle(style.ActiveStyleBase).Focused(true).WithHighlightedRow(0)
		case "esc", "left":
			if m.detailTableHasFocus {
				m.detailTableHasFocus = false
				m.detailTable = m.detailTable.Focused(false).WithBaseStyle(style.Base)
				m.deviceTable = m.deviceTable.Focused(true).WithBaseStyle(style.ActiveStyleBase)
			}
		}
	}
	if m.detailTableHasFocus {
		m.detailTable, cmd = m.detailTable.Update(msg)
	} else {
		m.deviceTable, cmd = m.deviceTable.Update(msg)
	}

	if m.deviceTable.HighlightedRow().Data["0"] != nil && m.selectedDevice != m.deviceTable.HighlightedRow().Data["0"] {
		m.selectedDevice = m.deviceTable.HighlightedRow().Data["0"].(string)

		m.detailTable = m.detailTable.WithRows(m.deviceProperties[m.selectedDevice].ToTableRows())
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
	m.deviceTable = m.deviceTable.WithPageSize(deviceTableHeight - 6)
	m.detailTable = m.detailTable.WithPageSize(detailsTableHeight - 6)
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)

	return m, tea.Batch(detailCMD, cmd)
}

func (m OptimizationDetailsPage) View() string {
	return m.deviceTable.View() + "\n" +
		wordwrap.String(m.item.Description, m.GetWidth()) + "\n" +
		m.detailTable.View() + "\n" +
		m.statusBar.View()
}

func (m OptimizationDetailsPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
