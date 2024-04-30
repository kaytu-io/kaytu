package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

var bold = lipgloss.NewStyle().Bold(true)
var changeFrom = lipgloss.NewStyle().Background(lipgloss.Color("88"))
var changeTo = lipgloss.NewStyle().Background(lipgloss.Color("28"))

var (
	styleBase = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("238")).
			Align(lipgloss.Left)
	activeStyleBase = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("248")).
			Align(lipgloss.Left)
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

type Ec2InstanceDetail struct {
	item             OptimizationItem
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

func PFloat64ToString(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *v)
}

func Percentage(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f%%", *v)
}

func PNetworkThroughputMbps(v *float64) string {
	if v == nil {
		return ""
	}
	vv := *v / 1000000 * 8
	return fmt.Sprintf("%.2f Mbps", vv)
}

func PNetworkThroughputMBps(v *float64) string {
	if v == nil {
		return ""
	}
	vv := *v / 1000000
	return fmt.Sprintf("%.2f MB/s", vv)
}

func NetworkThroughputMbps(v float64) string {
	return fmt.Sprintf("%.2f Mbps", v/1000000.0)
}

func PInt32ToString(v *int32) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func SizeByteToGB(v *int32) string {
	if v == nil {
		return ""
	}
	vv := *v // / 1000000000
	return fmt.Sprintf("%d GB", vv)
}

func ExtractProperties(item OptimizationItem) map[string]Rows {
	ifRecommendationExists := func(f func() string) string {
		if item.Wastage.RightSizing.Recommended != nil {
			return f()
		}
		return ""
	}

	res := map[string]Rows{
		*item.Instance.InstanceId: {
			{
				"",
				"",
				bold.Render("Average"),
				bold.Render("Max"),
				"",
			},
			{
				bold.Render("Region"),
				item.Region,
				"",
				"",
				item.Region,
			},
			{
				bold.Render("Instance Size"),
				string(item.Instance.InstanceType),
				"",
				"",
				ifRecommendationExists(func() string {
					return item.Wastage.RightSizing.Recommended.InstanceType
				}),
			},
			{
				bold.Render("Compute"),
				"",
				"",
				"",
				"",
			},
			{
				"  vCPU",
				fmt.Sprintf("%d", item.Wastage.RightSizing.Current.VCPU),
				Percentage(item.Wastage.RightSizing.VCPU.Avg),
				Percentage(item.Wastage.RightSizing.VCPU.Max),
				ifRecommendationExists(func() string {
					return fmt.Sprintf("%d", item.Wastage.RightSizing.Recommended.VCPU)
				}),
			},
			{
				"  Processor(s)",
				item.Wastage.RightSizing.Current.Processor,
				"",
				"",
				ifRecommendationExists(func() string {
					return item.Wastage.RightSizing.Recommended.Processor
				}),
			},
			{
				"  Architecture",
				item.Wastage.RightSizing.Current.Architecture,
				"",
				"",
				ifRecommendationExists(func() string {
					return item.Wastage.RightSizing.Recommended.Architecture
				}),
			},
			{
				"  Memory",
				fmt.Sprintf("%.1f GiB", item.Wastage.RightSizing.Current.Memory),
				Percentage(item.Wastage.RightSizing.Memory.Avg),
				Percentage(item.Wastage.RightSizing.Memory.Max),
				ifRecommendationExists(func() string {
					return fmt.Sprintf("%.1f GiB", item.Wastage.RightSizing.Recommended.Memory)
				}),
			},
			{
				bold.Render("EBS Bandwidth"),
				fmt.Sprintf("%s", item.Wastage.RightSizing.Current.EBSBandwidth),
				PNetworkThroughputMbps(item.Wastage.RightSizing.EBSBandwidth.Avg),
				PNetworkThroughputMbps(item.Wastage.RightSizing.EBSBandwidth.Max),
				ifRecommendationExists(func() string {
					return fmt.Sprintf("%s", item.Wastage.RightSizing.Recommended.EBSBandwidth)
				}),
			},
			{
				bold.Render("Network Performance"),
				"",
				"",
				"",
				"",
			},
			{
				"  Throughput",
				fmt.Sprintf("%s", item.Wastage.RightSizing.Current.NetworkThroughput),
				PNetworkThroughputMbps(item.Wastage.RightSizing.NetworkThroughput.Avg),
				PNetworkThroughputMbps(item.Wastage.RightSizing.NetworkThroughput.Max),
				ifRecommendationExists(func() string {
					return fmt.Sprintf("%s", item.Wastage.RightSizing.Recommended.NetworkThroughput)
				}),
			},
			{
				"  ENA",
				fmt.Sprintf("%s", item.Wastage.RightSizing.Current.ENASupported),
				"",
				"",
				ifRecommendationExists(func() string {
					return fmt.Sprintf("%s", item.Wastage.RightSizing.Recommended.ENASupported)
				}),
			},
		},
	}

	for _, v := range item.Volumes {
		vid := hash.HashString(*v.VolumeId)
		ifVolumeRecommendationExists := func(f func() string) string {
			if item.Wastage.VolumeRightSizing[vid].Recommended != nil {
				return f()
			}
			return ""
		}

		res[*v.VolumeId] = Rows{
			{
				"",
				"",
				bold.Render("Average"),
				bold.Render("Max"),
				"",
			},
			{
				"  EBS Storage Tier",
				string(item.Wastage.VolumeRightSizing[vid].Current.Tier),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return string(item.Wastage.VolumeRightSizing[vid].Recommended.Tier)
				}),
			},
			{
				"  Volume Size (GB)",
				SizeByteToGB(item.Wastage.VolumeRightSizing[vid].Current.VolumeSize),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return SizeByteToGB(item.Wastage.VolumeRightSizing[vid].Recommended.VolumeSize)
				}),
			},
			{
				bold.Render("IOPS"),
				fmt.Sprintf("%d", item.Wastage.VolumeRightSizing[vid].Current.IOPS()),
				PFloat64ToString(item.Wastage.VolumeRightSizing[vid].IOPS.Avg),
				PFloat64ToString(item.Wastage.VolumeRightSizing[vid].IOPS.Max),
				ifVolumeRecommendationExists(func() string {
					return fmt.Sprintf("%d", item.Wastage.VolumeRightSizing[vid].Recommended.IOPS())
				}),
			},
			{
				"  Baseline IOPS",
				fmt.Sprintf("%d", item.Wastage.VolumeRightSizing[vid].Current.BaselineIOPS),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return fmt.Sprintf("%d", item.Wastage.VolumeRightSizing[vid].Recommended.BaselineIOPS)
				}),
			},
			{
				"  Provisioned IOPS",
				PInt32ToString(item.Wastage.VolumeRightSizing[vid].Current.ProvisionedIOPS),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return PInt32ToString(item.Wastage.VolumeRightSizing[vid].Recommended.ProvisionedIOPS)
				}),
			},
			{
				bold.Render("Throughput (MB/s)"),
				fmt.Sprintf("%.2f", item.Wastage.VolumeRightSizing[vid].Current.Throughput()),
				PNetworkThroughputMBps(item.Wastage.VolumeRightSizing[vid].Throughput.Avg),
				PNetworkThroughputMBps(item.Wastage.VolumeRightSizing[vid].Throughput.Max),
				ifVolumeRecommendationExists(func() string {
					return fmt.Sprintf("%.2f", item.Wastage.VolumeRightSizing[vid].Recommended.Throughput())
				}),
			},
			{
				"  Baseline Throughput",
				NetworkThroughputMbps(item.Wastage.VolumeRightSizing[vid].Current.BaselineThroughput),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return NetworkThroughputMbps(item.Wastage.VolumeRightSizing[vid].Recommended.BaselineThroughput)
				}),
			},
			{
				"  Provisioned Throughput",
				PNetworkThroughputMbps(item.Wastage.VolumeRightSizing[vid].Current.ProvisionedThroughput),
				"",
				"",
				ifVolumeRecommendationExists(func() string {
					return PNetworkThroughputMbps(item.Wastage.VolumeRightSizing[vid].Recommended.ProvisionedThroughput)
				}),
			},
		}
	}

	for deviceID, rows := range res {
		for idx, row := range rows {
			if row[1] != row[4] {
				row[1] = changeFrom.Render(row[1])
				row[4] = changeTo.Render(row[4])
			}
			rows[idx] = row
		}
		res[deviceID] = rows
	}
	return res
}

func NewEc2InstanceDetail(item OptimizationItem, close func()) *Ec2InstanceDetail {
	ifRecommendationExists := func(f func() string) string {
		if item.Wastage.RightSizing.Recommended != nil {
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

	deviceRows := Rows{
		{
			*item.Instance.InstanceId,
			"EC2 Instance",
			"730 hours",
			fmt.Sprintf("$%.2f", item.Wastage.RightSizing.Current.Cost),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("$%.2f", item.Wastage.RightSizing.Recommended.Cost)
			}),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("$%.2f", item.Wastage.RightSizing.Current.Cost-item.Wastage.RightSizing.Recommended.Cost)
			}),
		},
	}
	for _, v := range item.Instance.BlockDeviceMappings {
		ifRecommendationExists := func(f func() string) string {
			if item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Recommended != nil {
				return f()
			}
			return ""
		}

		saving := 0.0
		if item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Recommended != nil {
			saving = item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Current.Cost - item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Recommended.Cost
		}
		deviceRows = append(deviceRows, Row{
			*v.Ebs.VolumeId,
			"EBS Volume",
			"730 hours",
			fmt.Sprintf("$%.2f", item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Current.Cost),
			ifRecommendationExists(func() string {
				return fmt.Sprintf("$%.2f", item.Wastage.VolumeRightSizing[hash.HashString(*v.Ebs.VolumeId)].Recommended.Cost)
			}),
			fmt.Sprintf("$%.2f", saving),
		})
	}

	days := "7"
	for _, p := range item.Preferences {
		if p.Key == "ObservabilityTimePeriod" && p.Value != nil {
			days = *p.Value
		}
	}
	detailColumns := []table.Column{
		table.NewColumn("0", "", 30),
		table.NewColumn("1", "Current", 30),
		table.NewColumn("2", fmt.Sprintf("%s day usage", days), 15),
		table.NewColumn("3", "", 15),
		table.NewColumn("4", "Recommendation", 30),
	}

	model := Ec2InstanceDetail{
		item:  item,
		close: close,
		detailTable: table.New(detailColumns).
			WithFooterVisibility(false).
			WithPageSize(1).
			WithBaseStyle(styleBase).BorderRounded(),
		deviceTable: table.New(deviceColumns).
			WithFooterVisibility(false).
			WithRows(deviceRows.ToTableRows()).
			Focused(true).
			WithPageSize(len(deviceRows)).
			WithBaseStyle(activeStyleBase).BorderRounded(),
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

func (m *Ec2InstanceDetail) Init() tea.Cmd { return nil }

func (m *Ec2InstanceDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.deviceTable = m.deviceTable.WithBaseStyle(styleBase)
			m.detailTable = m.detailTable.WithBaseStyle(activeStyleBase).Focused(true).WithHighlightedRow(0)
		case "esc", "left":
			if m.detailTableHasFocus {
				m.detailTableHasFocus = false
				m.detailTable = m.detailTable.Focused(false).WithBaseStyle(styleBase)
				m.deviceTable = m.deviceTable.Focused(true).WithBaseStyle(activeStyleBase)
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

func (m *Ec2InstanceDetail) View() string {
	return m.deviceTable.View() + "\n" +
		wordwrap.String(m.item.Wastage.RightSizing.Description, m.width) + "\n" +
		m.detailTable.View() + "\n" +
		m.help.String()
}

func (m *Ec2InstanceDetail) IsResponsive() bool {
	return m.height >= m.MinHeight()
}

func (m *Ec2InstanceDetail) SetHeight(height int) {
	l := strings.Count(wordwrap.String(m.item.Wastage.RightSizing.Description, m.width), "\n") + 1
	m.height = height
	m.help.SetHeight(m.height - (m.detailTable.TotalRows() + 4 + m.deviceTable.TotalRows() + 4 + l))
}

func (m *Ec2InstanceDetail) MinHeight() int {
	l := strings.Count(wordwrap.String(m.item.Wastage.RightSizing.Description, m.width), "\n") + 1
	return m.detailTable.TotalRows() + 4 + m.deviceTable.TotalRows() + 4 + m.help.MinHeight() + l
}
