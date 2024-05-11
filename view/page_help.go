package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view/responsive"
	"strings"
)

type HelpPage struct {
	helpController *controller.Help

	responsive.ResponsiveView
}

func NewHelpPage(helpController *controller.Help) HelpPage {
	return HelpPage{
		helpController: helpController,
	}
}

func (m HelpPage) OnClose() Page {
	return m
}
func (m HelpPage) OnOpen() Page {
	return m
}

func (m HelpPage) Init() tea.Cmd { return nil }

func (m HelpPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m HelpPage) View() string {
	var lines []string
	for _, line := range m.helpController.Help() {
		lines = append(lines, fmt.Sprintf(" %s", line))
	}

	return "\n" + style.HelpStyle.Render(strings.Join(lines, "\n")) + "\n"
}
func (m HelpPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
