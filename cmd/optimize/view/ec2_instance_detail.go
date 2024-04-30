package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

var bold = lipgloss.NewStyle().Bold(true)

type Ec2InstanceDetail struct {
	item             OptimizationItem
	close            func()
	deviceTable      table.Model
	detailTable      table.Model
	deviceProperties map[string][]table.Row
	width            int
	height           int
	selectedDevice   string
	help             HelpView
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

func ExtractProperties(item OptimizationItem) map[string][]table.Row {
	ifRecommendationExists := func(f func() string) string {
		if item.Wastage.RightSizing.Recommended != nil {
			return f()
		}
		return ""
	}

	res := map[string][]table.Row{
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
				"  License Cost",
				fmt.Sprintf("$%.2f", item.Wastage.RightSizing.Current.LicensePrice),
				"",
				"",
				ifRecommendationExists(func() string {
					return fmt.Sprintf("$%.2f", item.Wastage.RightSizing.Recommended.LicensePrice)
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

		res[*v.VolumeId] = []table.Row{
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
		{Title: "DeviceID", Width: 30},
		{Title: "ResourceType", Width: 20},
		{Title: "Runtime", Width: 13},
		{Title: "Current Cost", Width: 20},
		{Title: "Right sized Cost", Width: 20},
		{Title: "Savings", Width: 20},
	}
	deviceRows := []table.Row{
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
		deviceRows = append(deviceRows, table.Row{
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
		{Title: "", Width: 30},
		{Title: "Current", Width: 20},
		{Title: fmt.Sprintf("%s day usage", days), Width: 15},
		{Title: "", Width: 15},
		{Title: "Recommendation", Width: 30},
	}

	model := Ec2InstanceDetail{
		item:  item,
		close: close,
		detailTable: table.New(
			table.WithColumns(detailColumns),
			table.WithFocused(false),
			table.WithHeight(1),
		),
		deviceTable: table.New(
			table.WithColumns(deviceColumns),
			table.WithRows(deviceRows),
			table.WithFocused(true),
			table.WithHeight(len(deviceRows)),
		),
	}

	detailStyle := table.DefaultStyles()
	detailStyle.Header = detailStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	detailStyle.Selected = lipgloss.NewStyle()

	deviceStyle := table.DefaultStyles()
	deviceStyle.Header = deviceStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	deviceStyle.Selected = deviceStyle.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	model.detailTable.SetStyles(detailStyle)
	model.deviceTable.SetStyles(deviceStyle)
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
		m.deviceTable.SetWidth(m.width)
		m.detailTable.SetWidth(m.width)
		m.SetHeight(m.height)
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "left":
			m.close()
		}
	}
	m.deviceTable, cmd = m.deviceTable.Update(msg)
	if m.deviceTable.SelectedRow() != nil {
		if m.selectedDevice != m.deviceTable.SelectedRow()[0] {
			m.selectedDevice = m.deviceTable.SelectedRow()[0]
			m.detailTable.SetRows(m.deviceProperties[m.selectedDevice])
			m.detailTable.SetHeight(len(m.deviceProperties[m.selectedDevice]))
			m.SetHeight(m.height)
		}
	}
	//m.detailTable, detailCMD = m.detailTable.Update(msg)
	return m, tea.Batch(detailCMD, cmd)
}

func (m *Ec2InstanceDetail) View() string {
	return baseStyle.Render(m.deviceTable.View()) + "\n" +
		wordwrap.String(m.item.Wastage.RightSizing.Description, m.width) + "\n" +
		baseStyle.Render(m.detailTable.View()) + "\n" +
		m.help.String()
}

func (m *Ec2InstanceDetail) IsResponsive() bool {
	return m.height >= m.MinHeight()
}

func (m *Ec2InstanceDetail) SetHeight(height int) {
	l := strings.Count(wordwrap.String(m.item.Wastage.RightSizing.Description, m.width), "\n")
	m.height = height
	m.help.SetHeight(m.height - (len(m.detailTable.Rows()) + 4 + len(m.deviceTable.Rows()) + 4 + l))
}

func (m *Ec2InstanceDetail) MinHeight() int {
	l := strings.Count(wordwrap.String(m.item.Wastage.RightSizing.Description, m.width), "\n")
	return len(m.detailTable.Rows()) + 4 + len(m.deviceTable.Rows()) + 4 + m.help.MinHeight() + l
}
