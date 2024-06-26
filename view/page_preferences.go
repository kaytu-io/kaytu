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

type PreferencesPage[T golang.OptimizationItem | golang.ChartOptimizationItem] struct {
	focused int
	err     error

	serviceList     []string
	serviceIdx      int
	items           []*PreferenceItem
	visibleStartIdx int

	helpController *controller.Help
	optimizations  *controller.Optimizations[T]
	statusBar      StatusBarView
	responsive.ResponsiveView
}

func NewPreferencesConfiguration[T golang.OptimizationItem | golang.ChartOptimizationItem](
	helpController *controller.Help,
	optimizations *controller.Optimizations[T],
	statusBar StatusBarView,
) PreferencesPage[T] {
	return PreferencesPage[T]{
		helpController: helpController,
		optimizations:  optimizations,
		statusBar:      statusBar,
	}
}

func (m PreferencesPage[T]) OnOpen() Page {
	m.visibleStartIdx = 0
	var preferences []*golang.PreferenceItem
	if selectedItem := m.optimizations.SelectedItem(); selectedItem != nil {
		switch any(selectedItem).(type) {
		case *golang.OptimizationItem:
			selectedItem := any(selectedItem).(*golang.OptimizationItem)
			preferences = selectedItem.Preferences
		case *golang.ChartOptimizationItem:
			selectedItem := any(selectedItem).(*golang.ChartOptimizationItem)
			preferences = selectedItem.GetPreferences()
		}
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

func (m PreferencesPage[T]) OnClose() Page {
	selectedItem := m.optimizations.SelectedItem()
	if selectedItem == nil {
		for _, selectedItem := range m.optimizations.Items() {
			switch castedSelectedItem := any(selectedItem).(type) {
			case *golang.OptimizationItem:
				if castedSelectedItem == nil {
					continue
				}
				if castedSelectedItem.Skipped || castedSelectedItem.LazyLoadingEnabled {
					continue
				}
				var prefs []*golang.PreferenceItem
				for _, item := range m.items {
					prefs = append(prefs, item.pref)
				}
				castedSelectedItem.Preferences = prefs
				castedSelectedItem.Loading = true
				var a = any(*castedSelectedItem).(T)
				m.optimizations.SendItem(&a)
				m.optimizations.ReEvaluate(castedSelectedItem.Id, prefs)
			case *golang.ChartOptimizationItem:
				if castedSelectedItem == nil {
					continue
				}
				if castedSelectedItem.GetSkipped() || castedSelectedItem.GetLazyLoadingEnabled() {
					continue
				}
				var prefs []*golang.PreferenceItem
				for _, item := range m.items {
					prefs = append(prefs, item.pref)
				}
				castedSelectedItem.Preferences = prefs
				castedSelectedItem.Loading = true
				var a = any(*castedSelectedItem).(T)
				m.optimizations.SendItem(&a)
				m.optimizations.ReEvaluate(castedSelectedItem.GetOverviewChartRow().GetRowId(), prefs)
			}
		}
	} else {
		var prefs []*golang.PreferenceItem
		for _, item := range m.items {
			prefs = append(prefs, item.pref)
		}
		switch castedSelectedItem := any(selectedItem).(type) {
		case *golang.OptimizationItem:
			castedSelectedItem.Preferences = prefs
			castedSelectedItem.Loading = true
			m.optimizations.SendItem(selectedItem)
			m.optimizations.ReEvaluate(castedSelectedItem.Id, prefs)
		case *golang.ChartOptimizationItem:
			castedSelectedItem.Preferences = prefs
			castedSelectedItem.Loading = true
			m.optimizations.SendItem(selectedItem)
			m.optimizations.ReEvaluate(castedSelectedItem.GetOverviewChartRow().GetRowId(), prefs)
		}
	}
	return m
}

func (m PreferencesPage[T]) Init() tea.Cmd {
	return textinput.Blink
}

func (m PreferencesPage[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m PreferencesPage[T]) View() string {
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

func (m *PreferencesPage[T]) ChangeService(svc string) {
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
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if n < 0 {
		return errors.New("invalid number")
	}
	return nil
}

func (m *PreferencesPage[T]) fixVisibleStartIdx() {
	for m.focused < m.visibleStartIdx {
		m.visibleStartIdx--
	}

	visibleCount := m.GetHeight() - (4 + m.statusBar.Height())
	for m.focused >= m.visibleStartIdx+visibleCount {
		m.visibleStartIdx++
	}
}

func (m *PreferencesPage[T]) nextInput() {
	m.focused = (m.focused + 1) % len(m.items)
	if m.items[m.focused].hidden {
		m.nextInput()
	}
}

func (m *PreferencesPage[T]) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.items) - 1
	}
	if m.items[m.focused].hidden {
		m.prevInput()
	}
}

func (m PreferencesPage[T]) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
