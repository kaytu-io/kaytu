package view

import (
	"errors"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/style"
	preferences2 "github.com/kaytu-io/kaytu/preferences"
	"github.com/kaytu-io/kaytu/view/responsive"
	"strconv"
	"strings"
)

type (
	errMsg error
)

type PreferencesPage struct {
	focused int
	err     error

	serviceList     []string
	serviceIdx      int
	items           []*PreferenceItem
	visibleStartIdx int

	helpController *controller.Help
	optimizations  *controller.Optimizations
	statusBar      StatusBarView
	responsive.ResponsiveView
}

func NewPreferencesConfiguration(
	helpController *controller.Help,
	optimizations *controller.Optimizations,
	statusBar StatusBarView,
) PreferencesPage {
	return PreferencesPage{
		helpController: helpController,
		optimizations:  optimizations,
		statusBar:      statusBar,
	}
}

func (m PreferencesPage) OnOpen() Page {
	m.visibleStartIdx = 0
	var preferences []*golang.PreferenceItem
	if selectedItem := m.optimizations.SelectedItem(); selectedItem != nil {
		preferences = selectedItem.Preferences
	} else {
		preferences = preferences2.DefaultPreferences()
	}

	m.focused = 0
	m.items = nil
	m.serviceList = []string{"All"}
	for _, pref := range preferences {
		m.items = append(m.items, NewPreferenceItem(pref))

		exists := false
		for _, sv := range m.serviceList {
			if sv == pref.Service {
				exists = true
			}
		}

		if !exists {
			m.serviceList = append(m.serviceList, pref.Service)
		}
	}
	m.items[0].Focus()

	m.helpController.SetKeyMap([]string{
		"↑/↓: move",
		"enter: next field",
		"←/→: prev/next value (for fields with specific values)",
		//"ctrl + ←/→: prev/next change service filter",
		"esc: apply and exit",
		"tab: pin/unpin value to current resource",
		"ctrl+c: exit",
	})
	return m
}

func (m PreferencesPage) OnClose() Page {
	selectedItem := m.optimizations.SelectedItem()
	if selectedItem == nil {
		for _, selectedItem := range m.optimizations.Items() {
			if selectedItem.Skipped || selectedItem.LazyLoadingEnabled {
				continue
			}

			var prefs []*golang.PreferenceItem
			for _, item := range m.items {
				prefs = append(prefs, item.pref)
			}
			selectedItem.Preferences = prefs
			selectedItem.Loading = true
			m.optimizations.SendItem(selectedItem)
			m.optimizations.ReEvaluate(selectedItem.Id, prefs)
		}
	} else {
		var prefs []*golang.PreferenceItem
		for _, item := range m.items {
			prefs = append(prefs, item.pref)
		}
		selectedItem.Preferences = prefs
		selectedItem.Loading = true
		m.optimizations.SendItem(selectedItem)
		m.optimizations.ReEvaluate(selectedItem.Id, prefs)
	}
	return m
}

func (m PreferencesPage) Init() tea.Cmd {
	return textinput.Blink
}

func (m PreferencesPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.nextInput()
			m.fixVisibleStartIdx()
		case tea.KeyUp:
			m.prevInput()
			m.fixVisibleStartIdx()
		case tea.KeyDown:
			m.nextInput()
			m.fixVisibleStartIdx()
			//case tea.KeyCtrlRight:
			//	m.visibleStartIdx = 0
			//	m.serviceIdx++
			//	if m.serviceIdx >= len(m.serviceList) {
			//		m.serviceIdx = 0
			//	}
			//	m.ChangeService(m.serviceList[m.serviceIdx])
			//case tea.KeyCtrlLeft:
			//	m.visibleStartIdx = 0
			//	m.serviceIdx--
			//	if m.serviceIdx < 0 {
			//		m.serviceIdx = len(m.serviceList) - 1
			//	}
			//	m.ChangeService(m.serviceList[m.serviceIdx])
		}
		for i := range m.items {
			m.items[i].Blur()
		}
		m.items[m.focused].Focus()

	case errMsg:
		m.err = msg
		return m, nil
	}
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)

	_, cmd := m.items[m.focused].Update(msg)
	return m, cmd
}

func (m PreferencesPage) View() string {
	builder := strings.Builder{}

	builder.WriteString(style.SvcDisable.Render("Configure your preferences:"))
	//builder.WriteString(style.SvcDisable.Render("Configure your preferences on "))
	//for idx, svc := range m.serviceList {
	//	if idx == m.serviceIdx {
	//		builder.WriteString(style.SvcEnable.Render(fmt.Sprintf(" %s ", svc)))
	//	} else {
	//		builder.WriteString(style.SvcDisable.Render(fmt.Sprintf(" %s ", svc)))
	//	}
	//}
	//builder.WriteString(style.SvcDisable.Render("    "))

	visibleCount := m.GetHeight() - (4 + m.statusBar.Height())
	builder.WriteString("\n")
	if m.visibleStartIdx > 0 {
		builder.WriteString(" ⇡⇡⇡")
	}
	builder.WriteString("\n")

	var visibleItems []*PreferenceItem
	for _, pref := range m.items {
		if pref.hidden {
			continue
		}
		visibleItems = append(visibleItems, pref)
	}

	for _, pref := range visibleItems[m.visibleStartIdx:min(m.visibleStartIdx+visibleCount, len(visibleItems))] {
		builder.WriteString(pref.View())
	}

	if m.visibleStartIdx+visibleCount < len(visibleItems) {
		builder.WriteString(" ⇣⇣⇣")
	}

	builder.WriteString("\n\n")
	builder.WriteString(m.statusBar.View())
	return builder.String()
}

func (m *PreferencesPage) ChangeService(svc string) {
	if svc == "All" {
		for _, i := range m.items {
			i.hidden = false
			i.hideService = false
		}
		return
	}

	for _, i := range m.items {
		if i.pref.Service == svc {
			i.hidden = false
			i.hideService = true
		} else {
			i.hidden = true
			i.hideService = true
		}
	}
}

func pinnedValidator(s string) error {
	if s == "" {
		return nil
	}
	return errors.New("pinned")
}

func numberValidator(s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	if n < 0 {
		return errors.New("invalid number")
	}
	return nil
}

func (m *PreferencesPage) fixVisibleStartIdx() {
	for m.focused < m.visibleStartIdx {
		m.visibleStartIdx--
	}

	visibleCount := m.GetHeight() - (4 + m.statusBar.Height())
	for m.focused >= m.visibleStartIdx+visibleCount {
		m.visibleStartIdx++
	}
}

func (m *PreferencesPage) nextInput() {
	m.focused = (m.focused + 1) % len(m.items)
	if m.items[m.focused].hidden {
		m.nextInput()
	}
}

func (m *PreferencesPage) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.items) - 1
	}
	if m.items[m.focused].hidden {
		m.prevInput()
	}
}

func (m PreferencesPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
