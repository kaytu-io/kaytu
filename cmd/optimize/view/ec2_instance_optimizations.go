package view

import (
	"fmt"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var costStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
var savingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

type OptimizationItem struct {
	Instance            types.Instance
	Volumes             []types.Volume
	Region              string
	OptimizationLoading bool
	Skipped             bool
	SkipReason          *string

	Preferences   []preferences2.PreferenceItem
	Wastage       wastage.EC2InstanceWastageResponse
	Metrics       map[string][]types2.Datapoint
	VolumeMetrics map[string]map[string][]types2.Datapoint
}

type Ec2InstanceOptimizations struct {
	itemsChan chan OptimizationItem

	table table.Model
	items []OptimizationItem
	help  HelpView

	detailsPage *Ec2InstanceDetail
	prefConf    *PreferencesConfiguration

	clearScreen  bool
	instanceChan chan OptimizationItem

	Width  int
	height int

	tableHeight int
}

func NewEC2InstanceOptimizations(instanceChan chan OptimizationItem) *Ec2InstanceOptimizations {
	columns := []table.Column{
		table.NewColumn("0", "Instance Id", 23),
		table.NewColumn("1", "Instance Name", 23),
		table.NewColumn("2", "Instance Type", 15),
		table.NewColumn("3", "Region", 15),
		table.NewColumn("4", "Platform", 15),
		table.NewColumn("5", "Total Saving (Monthly)", 40),
		table.NewColumn("6", "", 1),
	}
	t := table.New(columns).
		Focused(true).
		WithPageSize(10).
		WithBaseStyle(activeStyleBase).
		BorderRounded()

	return &Ec2InstanceOptimizations{
		itemsChan: make(chan OptimizationItem, 1000),
		table:     t,
		items:     nil,
		help: HelpView{
			lines: []string{
				"↑/↓: move",
				"enter/→: see details",
				"p: change preferences for one item",
				"P: change preferences for all items",
				"q/ctrl+c: exit",
			},
		},
		instanceChan: instanceChan,
	}
}

func (m *Ec2InstanceOptimizations) Init() tea.Cmd { return tickCmdWithDuration(time.Millisecond * 50) }

func (m *Ec2InstanceOptimizations) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	}

	if m.detailsPage != nil {
		_, cmd := m.detailsPage.Update(msg)
		return m, tea.Batch(cmd, tickCmdWithDuration(time.Millisecond*50))
	}
	if m.prefConf != nil {
		_, cmd := m.prefConf.Update(msg)
		return m, tea.Batch(cmd, tickCmdWithDuration(time.Millisecond*50))
	}

	var cmd, initCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateResponsive()

	case tickMsg:
		for {
			nothingToAdd := false
			select {
			case newItem := <-m.itemsChan:
				updated := false
				for idx, i := range m.items {
					if *newItem.Instance.InstanceId == *i.Instance.InstanceId {
						m.items[idx] = newItem
						updated = true
						break
					}
				}
				if !updated {
					m.items = append(m.items, newItem)
				}

				var rows Rows
				for _, i := range m.items {
					platform := ""
					if i.Instance.PlatformDetails != nil {
						platform = *i.Instance.PlatformDetails
					}

					totalSaving := 0.0
					if i.Wastage.RightSizing.Recommended != nil {
						totalSaving = i.Wastage.RightSizing.Current.Cost - i.Wastage.RightSizing.Recommended.Cost
						for _, s := range i.Wastage.VolumeRightSizing {
							if s.Recommended != nil {
								totalSaving += s.Current.Cost - s.Recommended.Cost
							}
						}
					}

					name := ""
					for _, t := range i.Instance.Tags {
						if t.Key != nil && strings.ToLower(*t.Key) == "name" && t.Value != nil {
							name = *t.Value
						}
					}
					if name == "" {
						name = *i.Instance.InstanceId
					}

					row := Row{
						*i.Instance.InstanceId,
						name,
						string(i.Instance.InstanceType),
						i.Region,
						platform,
						fmt.Sprintf("$%.2f", totalSaving),
					}
					if i.OptimizationLoading {
						row[5] = "loading"
					} else if i.Skipped {
						row[5] = "skipped"
						if i.SkipReason != nil {
							row[5] += " - " + *i.SkipReason
						}
					}
					row = append(row, "→")
					rows = append(rows, row)
				}
				m.table = m.table.WithRows(rows.ToTableRows())
			default:
				nothingToAdd = true
			}
			if nothingToAdd {
				break
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "p":
			if m.table.TotalRows() == 0 {
				break
			}
			selectedInstanceID := m.table.HighlightedRow().Data["0"]
			for _, i := range m.items {
				if selectedInstanceID == *i.Instance.InstanceId {
					m.prefConf = NewPreferencesConfiguration(i.Preferences, func(items []preferences2.PreferenceItem) {
						i.Preferences = items
						i.OptimizationLoading = true
						m.itemsChan <- i
						m.prefConf = nil
						m.clearScreen = true
						// re-evaluate
						m.instanceChan <- i
						m.UpdateResponsive()
					}, m.Width)
					initCmd = m.prefConf.Init()
					break
				}
			}
			m.UpdateResponsive()
		case "P":
			if m.table.TotalRows() == 0 {
				break
			}

			m.prefConf = NewPreferencesConfiguration(preferences2.DefaultPreferences(), func(items []preferences2.PreferenceItem) {
				for _, i := range m.items {
					i.Preferences = items
					i.OptimizationLoading = true
					m.itemsChan <- i
					m.instanceChan <- i
				}
				m.prefConf = nil
				m.clearScreen = true
				m.UpdateResponsive()
			}, m.Width)
			initCmd = m.prefConf.Init()
			m.UpdateResponsive()
		case "enter", "right":
			if m.table.TotalRows() == 0 {
				break
			}

			selectedInstanceID := m.table.HighlightedRow().Data["0"]
			for _, i := range m.items {
				if selectedInstanceID == *i.Instance.InstanceId {
					m.detailsPage = NewEc2InstanceDetail(i, func() {
						m.detailsPage = nil
						m.UpdateResponsive()
					})
					initCmd = m.detailsPage.Init()
					m.detailsPage.width = m.Width
					break
				}
			}
			m.UpdateResponsive()
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmd = tea.Batch(cmd, tickCmdWithDuration(time.Millisecond*50))
	if initCmd != nil {
		cmd = tea.Batch(cmd, initCmd)
	}
	return m, cmd
}

func (m *Ec2InstanceOptimizations) View() string {
	if m.clearScreen {
		m.clearScreen = false
		return ""
	}
	if m.detailsPage != nil {
		return m.detailsPage.View()
	}
	if m.prefConf != nil {
		return m.prefConf.View()
	}

	totalCost := 0.0
	savings := 0.0
	for _, i := range m.items {
		totalCost += i.Wastage.RightSizing.Current.Cost
		if i.Wastage.RightSizing.Recommended != nil {
			savings += i.Wastage.RightSizing.Current.Cost - i.Wastage.RightSizing.Recommended.Cost
		}

		for _, v := range i.Wastage.VolumeRightSizing {
			totalCost += v.Current.Cost
		}
		for _, v := range i.Wastage.VolumeRightSizing {
			if v.Recommended != nil {
				savings += v.Current.Cost - v.Recommended.Cost
			}
		}
	}

	return fmt.Sprintf("Current runtime cost: %s, Savings: %s\n%s\n%s",
		costStyle.Render(fmt.Sprintf("$%.2f", totalCost)), savingStyle.Render(fmt.Sprintf("$%.2f", savings)),
		m.table.View(),
		m.help.String())
}

func (m *Ec2InstanceOptimizations) SendItem(item OptimizationItem) {
	m.itemsChan <- item
}

func (m *Ec2InstanceOptimizations) UpdateResponsive() {
	defer func() {
		m.table = m.table.WithPageSize(m.tableHeight - 7)
		if m.prefConf != nil {
			m.prefConf.SetHeight(m.tableHeight)
		}
		if m.detailsPage != nil {
			m.detailsPage.SetHeight(m.tableHeight)
		}
	}()

	if m.prefConf != nil || m.detailsPage != nil {
		m.tableHeight = m.height
		return
	}

	m.tableHeight = 8
	m.help.SetHeight(m.help.MinHeight())

	checkResponsive := func() bool {
		return m.height >= m.help.height+m.tableHeight && m.help.IsResponsive() && m.tableHeight >= 7
	}

	if !checkResponsive() {
		return // nothing to do
	}

	for m.tableHeight < 11 {
		m.tableHeight++
		if !checkResponsive() {
			m.tableHeight--
			return
		}
	}

	for m.help.height < m.help.MaxHeight() {
		m.help.SetHeight(m.help.height + 1)
		if !checkResponsive() {
			m.help.SetHeight(m.help.height - 1)
			return
		}
	}

	for m.tableHeight < 30 {
		m.tableHeight++
		if !checkResponsive() {
			m.tableHeight--
			return
		}
	}
}

func (m *Ec2InstanceOptimizations) SetHeight(height int) {
	m.height = height
	m.UpdateResponsive()
}

func (m *Ec2InstanceOptimizations) MinHeight() int {
	if m.prefConf != nil {
		return m.prefConf.MinHeight()
	}
	if m.detailsPage != nil {
		return m.detailsPage.MinHeight()
	}
	return m.help.MinHeight() + 7 + 1
}

func (m *Ec2InstanceOptimizations) PreferredMinHeight() int {
	return 15
}

func (m *Ec2InstanceOptimizations) MaxHeight() int {
	return m.help.MaxHeight() + 30
}

func (m *Ec2InstanceOptimizations) IsResponsive() bool {
	if m.prefConf != nil && !m.prefConf.IsResponsive() {
		return false
	}
	if m.detailsPage != nil && !m.detailsPage.IsResponsive() {
		return false
	}
	return m.height >= m.MinHeight()
}
