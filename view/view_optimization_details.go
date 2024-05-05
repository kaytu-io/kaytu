package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
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

type OptimizationDetailsView struct {
	item             *golang.OptimizationItem
	close            func()
	deviceTable      table.Model
	detailTable      table.Model
	deviceProperties map[string]Rows
	width            int
	height           int
	selectedDevice   string
	help             HelpView

	detailTableHasFocus bool
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

func NewOptimizationDetailsView(item *golang.OptimizationItem, close func()) *OptimizationDetailsView {
	ifRecommendationExists := func(f func() string) string {
		if !item.Loading && !item.Skipped {
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
			days = p.Value.String()
		}
	}
	detailColumns := []table.Column{
		table.NewColumn("0", "", 30),
		table.NewColumn("1", "Current", 30),
		table.NewColumn("2", fmt.Sprintf("%s day usage", days), 15),
		table.NewColumn("3", "", 15),
		table.NewColumn("4", "Recommendation", 30),
	}

	model := OptimizationDetailsView{
		item:  item,
		close: close,
		detailTable: table.New(detailColumns).
			WithFooterVisibility(false).
			WithPageSize(1).
			WithBaseStyle(style.Base).BorderRounded(),
		deviceTable: table.New(deviceColumns).
			WithFooterVisibility(false).
			WithRows(deviceRows.ToTableRows()).
			Focused(true).
			WithPageSize(len(deviceRows)).
			WithBaseStyle(style.ActiveStyleBase).BorderRounded(),
	}

	model.deviceProperties = ExtractProperties(item)
	model.help = HelpView{
		lines: []string{
			"↑/↓: move",
			"esc/←: back to ec2 instance list",
			"q/ctrl+c: exit",
		},
		height: 0,
	}
	return &model
}

func (m *OptimizationDetailsView) Init() tea.Cmd { return nil }

func (m *OptimizationDetailsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, detailCMD tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.detailTable = m.detailTable.WithMaxTotalWidth(m.width)
		m.deviceTable = m.deviceTable.WithMaxTotalWidth(m.width)
		m.SetHeight(m.height)
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
			} else {
				m.close()
			}
		}
	}
	if m.detailTableHasFocus {
		m.detailTable, cmd = m.detailTable.Update(msg)
	} else {
		m.deviceTable, cmd = m.deviceTable.Update(msg)
	}

	if m.selectedDevice != m.deviceTable.HighlightedRow().Data["0"] {
		m.selectedDevice = m.deviceTable.HighlightedRow().Data["0"].(string)

		m.detailTable = m.detailTable.WithRows(m.deviceProperties[m.selectedDevice].ToTableRows()).WithPageSize(len(m.deviceProperties[m.selectedDevice]))
		m.SetHeight(m.height)
	}
	return m, tea.Batch(detailCMD, cmd)
}

func (m *OptimizationDetailsView) View() string {
	return m.deviceTable.View() + "\n" +
		wordwrap.String(m.item.Description, m.width) + "\n" +
		m.detailTable.View() + "\n" +
		m.help.String()
}

func (m *OptimizationDetailsView) IsResponsive() bool {
	return m.height >= m.MinHeight()
}

func (m *OptimizationDetailsView) SetHeight(height int) {
	l := strings.Count(wordwrap.String(m.item.Description, m.width), "\n") + 1
	m.height = height
	m.help.SetHeight(m.height - (m.detailTable.TotalRows() + 4 + m.deviceTable.TotalRows() + 4 + l))
}

func (m *OptimizationDetailsView) MinHeight() int {
	l := strings.Count(wordwrap.String(m.item.Description, m.width), "\n") + 1
	return m.detailTable.TotalRows() + 4 + m.deviceTable.TotalRows() + 4 + m.help.MinHeight() + l
}
