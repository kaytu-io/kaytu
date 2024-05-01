package view

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/optimize/style"
	"github.com/muesli/reflow/wordwrap"
	"sort"
	"strings"
	"sync"
)

type Job struct {
	ID             string
	Descrption     string
	FailureMessage string
	Done           bool
}

type JobsView struct {
	runningJobsMap map[string]string
	failedJobsMap  map[string]string

	runningJobs     []string
	moreRunningJobs bool
	failedJobs      []string
	moreFailedJobs  bool
	statusErr       string

	height int
	width  int

	jobMutex  sync.RWMutex
	jobChan   chan Job
	errorChan chan error
}

func NewJobsView() *JobsView {
	jobView := JobsView{
		runningJobsMap:  map[string]string{},
		failedJobsMap:   map[string]string{},
		runningJobs:     nil,
		moreRunningJobs: false,
		failedJobs:      nil,
		moreFailedJobs:  false,
		statusErr:       "",
		height:          0,
		width:           0,
		jobMutex:        sync.RWMutex{},
		jobChan:         make(chan Job, 10000),
		errorChan:       make(chan error, 10000),
	}
	go jobView.UpdateStatus()

	return &jobView
}

func (m *JobsView) Init() tea.Cmd {
	return nil
}

func (m *JobsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.jobMutex.RLock()
	m.runningJobs, m.moreRunningJobs = m.RunningJobs()
	m.failedJobs, m.moreFailedJobs = m.FailedJobs()
	m.jobMutex.RUnlock()

	return m, nil
}

func (m *JobsView) RunningJobs() ([]string, bool) {
	if len(m.runningJobsMap) == 0 {
		return nil, false
	}
	var res []string
	for _, v := range m.runningJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	count := 3
	if len(res) < 3 {
		count = len(res)
	}
	return res[:count], len(m.runningJobsMap) > 3
}

func (m *JobsView) FailedJobs() ([]string, bool) {
	if len(m.failedJobsMap) == 0 {
		return nil, false
	}
	var res []string
	for _, v := range m.failedJobsMap {
		res = append(res, v)
	}
	sort.Strings(res)
	count := 3
	if len(res) < 3 {
		count = len(res)
	}
	return res[:count], len(m.failedJobsMap) > 3
}

func (m *JobsView) UpdateStatus() {
	for {
		select {
		case job := <-m.jobChan:
			m.jobMutex.Lock()
			if !job.Done {
				m.runningJobsMap[job.ID] = job.Descrption
			} else {
				if _, ok := m.runningJobsMap[job.ID]; ok {
					delete(m.runningJobsMap, job.ID)
				}
			}
			if len(job.FailureMessage) > 0 {
				m.failedJobsMap[job.ID] = fmt.Sprintf("%s failed due to %s", job.Descrption, job.FailureMessage)
			}
			m.jobMutex.Unlock()

		case err := <-m.errorChan:
			m.statusErr = fmt.Sprintf("Failed due to %v", err)
		}
	}
}

func (m *JobsView) SetWidth(width int) {
	m.width = width
}

func (m *JobsView) SetHeight(height int) {
	m.height = height
}

func (m *JobsView) MinHeight() int {
	statusHeight := 0
	if len(m.statusErr) > 0 {
		statusHeight = strings.Count(wordwrap.String("  error: "+m.statusErr, m.width), "\n") + 1
	}
	return statusHeight + 1
}

func (m *JobsView) MaxHeight() int {
	statusHeight := 0
	if len(m.statusErr) > 0 {
		statusHeight = strings.Count(wordwrap.String("  error: "+m.statusErr, m.width), "\n") + 1
	}
	maxFailedLines := len(m.failedJobs)
	maxRunningLines := len(m.runningJobs)
	if m.moreRunningJobs {
		maxRunningLines++
	}
	if m.moreFailedJobs {
		maxFailedLines++
	}
	return statusHeight + maxRunningLines + maxFailedLines
}

func (m *JobsView) IsResponsive() bool {
	return m.height >= m.MinHeight()
}

func (m *JobsView) View() string {
	maxFailedLines := len(m.failedJobs)
	maxRunningLines := len(m.runningJobs)
	if m.moreRunningJobs {
		maxRunningLines++
	}
	if m.moreFailedJobs {
		maxFailedLines++
	}
	runningShowCount := 0
	failedShowCount := 0

	for runningShowCount+failedShowCount < m.height {
		if runningShowCount == 0 && runningShowCount < len(m.runningJobs) {
			runningShowCount++
			continue
		}

		if failedShowCount < len(m.failedJobs) {
			failedShowCount++
		} else if runningShowCount < len(m.runningJobs) {
			runningShowCount++
		} else if failedShowCount < maxFailedLines {
			failedShowCount++
		} else if runningShowCount < maxRunningLines {
			runningShowCount++
		} else {
			break
		}
	}

	var lines []string
	if runningShowCount > 0 && len(m.runningJobs) > 0 {
		for idx, v := range m.runningJobs {
			if runningShowCount == 0 {
				break
			}
			line := fmt.Sprintf("       - %s", v)
			if idx == 0 {
				line = fmt.Sprintf(" jobs: - %s", v)
			}
			lines = append(lines, wordwrap.String(line, m.width))
			runningShowCount--
		}
		if m.moreRunningJobs && runningShowCount > 0 {
			lines = append(lines, "       ...")
		}
	}
	if failedShowCount > 0 && len(m.failedJobs) > 0 {
		for idx, v := range m.failedJobs {
			if failedShowCount == 0 {
				break
			}
			line := fmt.Sprintf("         - %s", v)
			if idx == 0 {
				line = fmt.Sprintf(" failures: - %s", v)
			}
			lines = append(lines, style.ErrorStyle.Render(wordwrap.String(line, m.width)))
			failedShowCount--
		}
		if m.moreFailedJobs && failedShowCount > 0 {
			lines = append(lines, style.ErrorStyle.Render("       ..."))
		}
	}

	statusErr := ""
	if len(m.statusErr) > 0 {
		statusErr = style.ErrorStyle.Render(wordwrap.String("  error: "+m.statusErr, m.width)) + "\n"
	}

	return statusErr + strings.Join(lines, "\n")
}

func (m *JobsView) PublishError(err error) {
	m.errorChan <- err
}

func (m *JobsView) Publish(job Job) Job {
	m.jobChan <- job
	return job
}
