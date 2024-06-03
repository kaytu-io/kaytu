package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/style"
	"strings"
)

type StatusBarView struct {
	helpController *controller.Help
	jobsController *controller.Jobs
	initialization bool
	content        string
	width          int
}

func NewStatusBarView(JobsController *controller.Jobs, helpController *controller.Help) StatusBarView {
	return StatusBarView{jobsController: JobsController, helpController: helpController}
}

func (v StatusBarView) Init() tea.Cmd { return nil }
func (v StatusBarView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
	}

	failedJobs := v.jobsController.FailedJobs()
	runningCount, failedCount := len(v.jobsController.RunningJobs()), len(failedJobs)

	var status []string

	var helpLines []string
	w := 0

	if v.initialization {
		line := " initializing "
		w += len(line)
		helpLines = append(helpLines, style.JobsStatusStyle.Render(line))
	}

	if runningCount > 0 {
		line := fmt.Sprintf(" running jobs: %d ", runningCount)
		w += len(line)
		helpLines = append(helpLines, style.JobsStatusStyle.Render(line))
	}

	for idx, line := range v.helpController.Help() {
		line = fmt.Sprintf(" %s ", line)
		w += len(line)
		if w > v.width {
			helpLines = append(helpLines, "\n")
			w = 0
		}

		if idx%2 == 0 {
			helpLines = append(helpLines, style.InfoStatusStyle.Render(line))
		} else {
			helpLines = append(helpLines, style.InfoStatusStyle2.Render(line))
		}
	}
	status = append(status, strings.Join(helpLines, "")+"\n")

	if err := v.jobsController.GetError(); len(err) > 0 {
		status = append(status, style.ErrorStatusStyle.Render(strings.TrimSpace(err))+"\n")
	}
	if failedCount > 0 {
		status = append(status, style.ErrorStatusStyle.Render(fmt.Sprintf("failed job: %s, press ctrl+j to see more", failedJobs[0]))+"\n")
	}
	v.content = strings.Join(status, "")
	return v, nil
}
func (v StatusBarView) View() string {
	return v.content
}

func (v StatusBarView) Height() int {
	return strings.Count(v.content, "\n") + 1
}
