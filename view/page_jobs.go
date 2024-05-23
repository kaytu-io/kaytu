package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/view/responsive"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

type JobsPage struct {
	helpController *controller.Help
	jobController  *controller.Jobs
	statusBar      StatusBarView

	responsive.ResponsiveView
}

func NewJobsPage(jobController *controller.Jobs, helpController *controller.Help, statusBar StatusBarView) JobsPage {
	return JobsPage{
		jobController:  jobController,
		helpController: helpController,
		statusBar:      statusBar,
	}
}

func (m JobsPage) OnClose() Page {
	return m
}
func (m JobsPage) OnOpen() Page {
	m.helpController.SetKeyMap([]string{
		"esc: back to main menu",
		"q/ctrl+c: exit",
	})

	return m
}

func (m JobsPage) Init() tea.Cmd { return nil }

func (m JobsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newStatusBar, _ := m.statusBar.Update(msg)
	m.statusBar = newStatusBar.(StatusBarView)

	return m, nil
}

func (m JobsPage) View() string {
	runningJobs := m.jobController.RunningJobs()
	failedJobs := m.jobController.FailedJobs()

	var lines []string
	statusErr := ""
	if len(m.jobController.GetError()) > 0 {
		statusErr = style.ErrorStyle.Render(wordwrap.String("  error: "+m.jobController.GetError(), m.GetWidth())) + "\n"
	}

	for idx, v := range failedJobs {
		line := fmt.Sprintf("         - %s", v)
		if idx == 0 {
			line = fmt.Sprintf(" failures: - %s", v)
		}
		lines = append(lines, style.ErrorStyle.Render(wordwrap.String(line, m.GetWidth())))
	}

	for idx, v := range runningJobs {
		line := fmt.Sprintf("       - %s", v)
		if idx == 0 {
			line = fmt.Sprintf(" jobs: - %s", v)
		}
		lines = append(lines, wordwrap.String(line, m.GetWidth()))
	}

	if len(runningJobs) == 0 {
		lines = append(lines, " no running job")
	}

	return "\n" + statusErr + strings.Join(lines, "\n") + "\n\n" +
		m.statusBar.View()
}
func (m JobsPage) SetResponsiveView(rv responsive.ResponsiveViewInterface) Page {
	m.ResponsiveView = rv.(responsive.ResponsiveView)
	return m
}
