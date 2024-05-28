package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/view/responsive"
	"strings"
)

type ContactUsPage struct {
	helpController *controller.Help

	responsive.ResponsiveView
}

func NewContactUsPage(helpController *controller.Help) ContactUsPage {
	return ContactUsPage{
		helpController: helpController,
	}
}

func (m ContactUsPage) OnClose() Page {
	return m
}
func (m ContactUsPage) OnOpen() Page {
	m.helpController.SetKeyMap([]string{
		"ctrl+c: exit",
	})
	return m
}

func (m ContactUsPage) Init() tea.Cmd { return nil }

func (m ContactUsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ContactUsPage) View() string {
	message := []string{utils.ContactUsMessage}
	var helpLines []string
	for idx, line := range m.helpController.Help() {
		line = fmt.Sprintf(" %s ", line)

		if idx%2 == 0 {
			helpLines = append(helpLines, style.InfoStatusStyle.Render(line))
		} else {
			helpLines = append(helpLines, style.InfoStatusStyle2.Render(line))
		}
	}
	message = append(message, "\n\n"+strings.Join(helpLines, "")+"\n")

	return strings.Join(message, "")
}
func (m ContactUsPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
