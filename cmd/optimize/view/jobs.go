package view

import (
	"fmt"
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

type JobsView struct {
	runningJobs     []string
	moreRunningJobs bool
	failedJobs      []string
	moreFailedJobs  bool

	height int
	width  int
}

func (h *JobsView) SetWidth(width int) {
	h.width = width
}

func (h *JobsView) SetHeight(height int) {
	h.height = height
}

func (h *JobsView) MinHeight() int {
	return 1
}

func (h *JobsView) MaxHeight() int {
	maxFailedLines := len(h.failedJobs)
	maxRunningLines := len(h.runningJobs)
	if h.moreRunningJobs {
		maxRunningLines++
	}
	if h.moreFailedJobs {
		maxFailedLines++
	}
	return maxRunningLines + maxFailedLines
}

func (h *JobsView) IsResponsive() bool {
	return h.height >= h.MinHeight()
}

func (h *JobsView) String() string {
	maxFailedLines := len(h.failedJobs)
	maxRunningLines := len(h.runningJobs)
	if h.moreRunningJobs {
		maxRunningLines++
	}
	if h.moreFailedJobs {
		maxFailedLines++
	}
	runningShowCount := 0
	failedShowCount := 0

	for runningShowCount+failedShowCount < h.height {
		if runningShowCount == 0 && runningShowCount < len(h.runningJobs) {
			runningShowCount++
			continue
		}

		if failedShowCount < len(h.failedJobs) {
			failedShowCount++
		} else if runningShowCount < len(h.runningJobs) {
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
	if runningShowCount > 0 && len(h.runningJobs) > 0 {
		for idx, v := range h.runningJobs {
			if runningShowCount == 0 {
				break
			}
			line := fmt.Sprintf("       - %s", v)
			if idx == 0 {
				line = fmt.Sprintf(" jobs: - %s", v)
			}
			lines = append(lines, wordwrap.String(line, h.width))
			runningShowCount--
		}
		if h.moreRunningJobs && runningShowCount > 0 {
			lines = append(lines, "       ...")
		}
	}
	if failedShowCount > 0 && len(h.failedJobs) > 0 {
		for idx, v := range h.failedJobs {
			if failedShowCount == 0 {
				break
			}
			line := fmt.Sprintf("         - %s", v)
			if idx == 0 {
				line = fmt.Sprintf(" failures: - %s", v)
			}
			lines = append(lines, errorStyle.Render(wordwrap.String(line, h.width)))
			failedShowCount--
		}
		if h.moreFailedJobs && failedShowCount > 0 {
			lines = append(lines, errorStyle.Render("       ..."))
		}
	}

	return strings.Join(lines, "\n")
}
