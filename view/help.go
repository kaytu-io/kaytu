package view

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/style"
	"math"
	"strings"
)

type HelpView struct {
	lines  []string
	height int
}

func (h *HelpView) String() string {
	if h.height == 0 {
		return ""
	}
	lines := h.lines

	joinCount := math.Ceil(float64(len(lines)) / float64(h.height))
	if joinCount > 1 {
		lines = h.joinLines(lines, int(joinCount))
	}

	var hlines []string
	for _, line := range lines {
		hlines = append(hlines, fmt.Sprintf("    %s", line))
	}

	prefix := ""
	suffix := ""
	if len(lines) < h.height {
		suffix = "\n"
		if len(lines)+1 < h.height {
			prefix = "\n"
		}
	}
	return prefix + style.HelpStyle.Render(strings.Join(hlines, "\n")) + suffix + "\n"
}

func (h *HelpView) joinLines(lines []string, n int) []string {
	var newLines []string

	idx := 0
	for idx < len(lines) {
		var currentLine []string
		for i := 0; i < n; i++ {
			if idx >= len(lines) {
				break
			}
			currentLine = append(currentLine, lines[idx])
			idx++
		}
		newLines = append(newLines, strings.Join(currentLine, " | "))
	}
	return newLines
}

func (h *HelpView) SetHeight(height int) {
	h.height = height
}

func (h *HelpView) MinHeight() int {
	return 0
}

func (h *HelpView) MaxHeight() int {
	return len(h.lines) + 2
}

func (h *HelpView) IsResponsive() bool {
	return h.height >= h.MinHeight()
}
