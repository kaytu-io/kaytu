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

type PremiumPage struct {
	helpController *controller.Help

	responsive.ResponsiveView
}

func NewPremiumPage(helpController *controller.Help) PremiumPage {
	return PremiumPage{
		helpController: helpController,
	}
}

func (m PremiumPage) OnClose() Page {
	return m
}
func (m PremiumPage) OnOpen() Page {
	m.helpController.SetKeyMap([]string{
		"ctrl+c: exit",
	})
	return m
}

func (m PremiumPage) Init() tea.Cmd { return nil }

func (m PremiumPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m PremiumPage) View() string {
	message := []string{fmt.Sprintf("You have reached the limit for this user and organization.\n"+
		"You need to buy premium to use unlimitted edition:\n"+
		"%s", utils.BookMeetingURL)}
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
func (m PremiumPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
