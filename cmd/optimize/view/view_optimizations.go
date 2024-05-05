package view

import (
	"fmt"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/style"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evertras/bubble-table/table"
)

type Property struct {
	Key string

	Current     string
	Average     string
	Max         string
	Recommended string
}

type Device struct {
	Properties []Property

	DeviceID       string
	ResourceType   string
	Runtime        string
	CurrentCost    float64
	RightSizedCost float64
}

type OptimizationItem struct {
	ID           string
	Name         string
	ResourceType string
	Region       string
	Platform     string

	Devices     []Device
	Preferences []preferences2.PreferenceItem
	Description string

	Loading    bool
	Skipped    bool
	SkipReason *string
}

type OptimizationsView struct {
	itemsChan chan OptimizationItem

	table table.Model
	items []OptimizationItem
	help  HelpView

	detailsPage *OptimizationDetailsView
	prefConf    *PreferencesConfiguration

	clearScreen       bool
	ec2ReEvaluateFunc func(id string, items []preferences2.PreferenceItem)
	rdsReEvaluateFunc func(id string, items []preferences2.PreferenceItem)

	Width  int
	height int

	tableHeight int
}

func NewOptimizationsView() *OptimizationsView {
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
		WithBaseStyle(style.ActiveStyleBase).
		BorderRounded()

	return &OptimizationsView{
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
	}
}

func (m *OptimizationsView) SetReEvaluateFunc(ec2Evaluate func(id string, items []preferences2.PreferenceItem), rdsEvaluate func(id string, items []preferences2.PreferenceItem)) {
	m.ec2ReEvaluateFunc = ec2Evaluate
	m.rdsReEvaluateFunc = rdsEvaluate
}

func (m *OptimizationsView) Init() tea.Cmd { return tickCmdWithDuration(time.Millisecond * 50) }

func (m *OptimizationsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					if newItem.ID == i.ID {
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
					totalSaving := 0.0
					if !i.Loading && !i.Skipped {
						for _, dev := range i.Devices {
							totalSaving += dev.CurrentCost - dev.RightSizedCost
						}
					}

					row := Row{
						i.ID,
						i.Name,
						i.ResourceType,
						i.Region,
						i.Platform,
						fmt.Sprintf("$%.2f", totalSaving),
					}
					if i.Loading {
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
		key := msg.String()
		switch key {
		case "q":
			return m, tea.Quit
		case "p":
			if m.table.TotalRows() == 0 {
				break
			}
			selectedInstanceID := m.table.HighlightedRow().Data["0"]
			for _, i := range m.items {
				if selectedInstanceID == i.ID {
					m.prefConf = NewPreferencesConfiguration(preferences2.FilterServices(i.Preferences, map[string]bool{"all": true}), func(items []preferences2.PreferenceItem) {
						i.Preferences = items
						i.Loading = true
						m.itemsChan <- i
						m.prefConf = nil
						m.clearScreen = true
						// re-evaluate
						switch i.ResourceType {
						case "EC2 Instance":
							m.ec2ReEvaluateFunc(i.ID, items)
						case "RDS Instance":
							m.rdsReEvaluateFunc(i.ID, items)
						}

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

			m.prefConf = NewPreferencesConfiguration(preferences2.DefaultPreferences(map[string]bool{"all": true}), func(items []preferences2.PreferenceItem) {
				for _, i := range m.items {
					i.Preferences = items
					i.Loading = true
					m.itemsChan <- i

					switch i.ResourceType {
					case "EC2 Instance":
						m.ec2ReEvaluateFunc(i.ID, items)
					case "RDS Instance":
						m.rdsReEvaluateFunc(i.ID, items)
					}
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
				if selectedInstanceID == i.ID {
					m.detailsPage = NewOptimizationDetailsView(i, func() {
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

func (m *OptimizationsView) View() string {
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
		for _, dev := range i.Devices {
			totalCost += dev.CurrentCost
			savings += dev.CurrentCost - dev.RightSizedCost
		}
	}

	return fmt.Sprintf("Current runtime cost: %s, Savings: %s\n%s\n%s",
		style.CostStyle.Render(fmt.Sprintf("$%.2f", totalCost)), style.SavingStyle.Render(fmt.Sprintf("$%.2f", savings)),
		m.table.View(),
		m.help.String())
}

func (m *OptimizationsView) SendItem(item OptimizationItem) {
	m.itemsChan <- item
}

func (m *OptimizationsView) UpdateResponsive() {
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

func (m *OptimizationsView) SetHeight(height int) {
	m.height = height
	m.UpdateResponsive()
}

func (m *OptimizationsView) MinHeight() int {
	if m.prefConf != nil {
		return m.prefConf.MinHeight()
	}
	if m.detailsPage != nil {
		return m.detailsPage.MinHeight()
	}
	return m.help.MinHeight() + 7 + 1
}

func (m *OptimizationsView) PreferredMinHeight() int {
	return 15
}

func (m *OptimizationsView) MaxHeight() int {
	return m.help.MaxHeight() + 30
}

func (m *OptimizationsView) IsResponsive() bool {
	if m.prefConf != nil && !m.prefConf.IsResponsive() {
		return false
	}
	if m.detailsPage != nil && !m.detailsPage.IsResponsive() {
		return false
	}
	return m.height >= m.MinHeight()
}
