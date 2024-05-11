package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/style"
	"strings"
)

type StatusBarView struct {
	jobsController *controller.Jobs
	content        string
}

func NewStatusBarView(JobsController *controller.Jobs) StatusBarView {
	return StatusBarView{jobsController: JobsController}
}

func (v StatusBarView) Init() tea.Cmd { return nil }
func (v StatusBarView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	runningCount, failedCount := len(v.jobsController.RunningJobs()), len(v.jobsController.FailedJobs())

	var status []string
	if err := v.jobsController.GetError(); len(err) > 0 {
		status = append(status, strings.TrimSpace(err)+"\n")
	}
	if runningCount > 0 {
		status = append(status, fmt.Sprintf("running jobs: %d", runningCount))
	}
	if failedCount > 0 {
		status = append(status, fmt.Sprintf("failed jobs: %d", failedCount))
	}
	if runningCount > 0 || failedCount > 0 {
		status = append(status, fmt.Sprintf("press ctrl+j to see list of jobs"))
	}

	status = append(status, fmt.Sprintf("press ctrl+h to see help page"))

	v.content = strings.Join(status, ", ")
	return v, nil
}
func (v StatusBarView) View() string {
	return style.StatusBarStyle.Render(v.content)
}

func (v StatusBarView) Height() int {
	return strings.Count(v.content, "\n") + 1
}
