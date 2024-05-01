package view

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/style"
	"strconv"
	"strings"
)

type (
	errMsg error
)

type PreferencesConfiguration struct {
	focused int
	err     error
	height  int
	width   int
	help    HelpView

	serviceList []string
	serviceIdx  int
	items       []*PreferenceItem
	close       func([]preferences2.PreferenceItem)
}

func NewPreferencesConfiguration(preferences []preferences2.PreferenceItem, close func([]preferences2.PreferenceItem), width int) *PreferencesConfiguration {
	var items []*PreferenceItem
	serviceList := []string{"All"}
	for _, pref := range preferences {
		items = append(items, NewPreferenceItem(pref))

		exists := false
		for _, sv := range serviceList {
			if sv == pref.Service {
				exists = true
			}
		}

		if !exists {
			serviceList = append(serviceList, pref.Service)
		}
	}
	items[0].Focus()
	return &PreferencesConfiguration{
		items:       items,
		close:       close,
		serviceList: serviceList,
		width:       width,
		help: HelpView{
			lines: []string{
				"↑/↓: move",
				"enter: next field",
				"←/→: prev/next value (for fields with specific values)",
				"ctrl + ←/→: prev/next change service filter",
				"esc: apply and exit",
				"tab: pin/unpin value to current ec2 instance",
				"ctrl+c: exit",
			},
			height: 0,
		},
	}
}

func (m *PreferencesConfiguration) Init() tea.Cmd { return textinput.Blink }

func (m *PreferencesConfiguration) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			var prefs []preferences2.PreferenceItem
			for _, item := range m.items {
				prefs = append(prefs, item.pref)
			}
			m.close(prefs)
			return m, nil
		case tea.KeyEnter:
			m.nextInput()
		case tea.KeyUp:
			m.prevInput()
		case tea.KeyDown:
			m.nextInput()
		case tea.KeyCtrlRight:
			m.serviceIdx++
			if m.serviceIdx >= len(m.serviceList) {
				m.serviceIdx = 0
			}
			m.ChangeService(m.serviceList[m.serviceIdx])
		case tea.KeyCtrlLeft:
			m.serviceIdx--
			if m.serviceIdx < 0 {
				m.serviceIdx = len(m.serviceList) - 1
			}
			m.ChangeService(m.serviceList[m.serviceIdx])
		}
		for i := range m.items {
			m.items[i].Blur()
		}
		m.items[m.focused].Focus()

	case errMsg:
		m.err = msg
		return m, nil
	}

	_, cmd := m.items[m.focused].Update(msg)
	return m, cmd
}

func (m *PreferencesConfiguration) View() string {
	builder := strings.Builder{}

	builder.WriteString(style.SvcDisable.Render("Configure your preferences: "))
	for idx, svc := range m.serviceList {
		if idx == m.serviceIdx {
			builder.WriteString(style.SvcEnable.Render(fmt.Sprintf(" %s ", svc)))
		} else {
			builder.WriteString(style.SvcDisable.Render(fmt.Sprintf(" %s ", svc)))
		}
	}
	builder.WriteString(style.SvcDisable.Render("    "))
	builder.WriteString("\n\n")

	for _, pref := range m.items {
		builder.WriteString(pref.View())
	}
	builder.WriteString(m.help.String())
	return builder.String()
}

func (m *PreferencesConfiguration) ChangeService(svc string) {
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

func (m *PreferencesConfiguration) nextInput() {
	m.focused = (m.focused + 1) % len(m.items)
	if m.items[m.focused].hidden {
		m.nextInput()
	}
}

func (m *PreferencesConfiguration) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.items) - 1
	}
	if m.items[m.focused].hidden {
		m.prevInput()
	}
}

func (m *PreferencesConfiguration) IsResponsive() bool {
	return m.height >= m.MinHeight()
}

func (m *PreferencesConfiguration) SetHeight(height int) {
	m.height = height
	m.help.SetHeight(m.height - (len(m.items) + 3))
}

func (m *PreferencesConfiguration) MinHeight() int {
	return len(m.items) + 3 + m.help.MinHeight()
}
