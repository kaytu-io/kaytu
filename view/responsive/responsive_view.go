package responsive

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ResponsiveViewInterface interface {
	IsResponsive() bool
	SetSize(msg tea.WindowSizeMsg) ResponsiveViewInterface

	GetHeight() int
	GetWidth() int
	getMinHeight() int
	getMinWidth() int
	getMaxHeight() int
	getMaxWidth() int
}

type ResponsiveView struct {
	height, minHeight, maxHeight int
	width, minWidth, maxWidth    int
	children                     []ResponsiveViewInterface
}

func (rv ResponsiveView) GetHeight() int {
	return rv.height
}
func (rv ResponsiveView) GetWidth() int {
	return rv.width
}
func (rv *ResponsiveView) SetSizeBound(minHeight, maxHeight, minWidth, maxWidth int) {
	rv.minHeight = minHeight
	rv.maxHeight = maxHeight
	rv.minWidth = minWidth
	rv.maxWidth = maxWidth
}
func (rv ResponsiveView) getMinHeight() int {
	return rv.minHeight
}
func (rv ResponsiveView) getMinWidth() int {
	return rv.minWidth
}
func (rv ResponsiveView) getMaxHeight() int {
	return rv.maxHeight
}
func (rv ResponsiveView) getMaxWidth() int {
	return rv.maxWidth
}

func (rv ResponsiveView) IsResponsive() bool {
	for _, child := range rv.children {
		if !child.IsResponsive() {
			return false
		}
	}
	return rv.height >= rv.minHeight && rv.width >= rv.minWidth
}

func (rv ResponsiveView) SetSize(msg tea.WindowSizeMsg) ResponsiveViewInterface {
	if rv.maxHeight == 0 {
		rv.maxHeight = 9999
	}
	if rv.maxWidth == 0 {
		rv.maxWidth = 9999
	}

	rv.height = max(min(msg.Height, rv.maxHeight), rv.minHeight)
	rv.width = max(min(msg.Width, rv.maxWidth), rv.minWidth)

	if len(rv.children) == 0 {
		return rv
	}

	canIncreaseChildrenWidth := func() bool {
		sumW := 0
		for _, child := range rv.children {
			sumW += child.GetWidth()
		}
		return sumW < rv.width
	}

	canIncreaseChildrenHeight := func() bool {
		sumH := 0
		for _, child := range rv.children {
			sumH += child.GetHeight()
		}
		return sumH < rv.height
	}

	for idx, child := range rv.children {
		rv.children[idx] = child.SetSize(tea.WindowSizeMsg{
			Width:  child.getMinWidth(),
			Height: child.getMinHeight(),
		})
	}

	idx := 0
	skip := 0
	for canIncreaseChildrenWidth() {
		if rv.children[idx].GetWidth() < rv.children[idx].getMaxWidth() {
			rv.children[idx] = rv.children[idx].SetSize(tea.WindowSizeMsg{
				Width:  rv.children[idx].GetWidth() + 1,
				Height: rv.children[idx].GetHeight(),
			})
			skip = 0
		} else {
			skip++
			if skip == len(rv.children) {
				break
			}
		}
		idx = (idx + 1) % len(rv.children)
	}

	idx = 0
	skip = 0
	for canIncreaseChildrenHeight() {
		if rv.children[idx].GetHeight() < rv.children[idx].getMaxHeight() {
			rv.children[idx] = rv.children[idx].SetSize(tea.WindowSizeMsg{
				Width:  rv.children[idx].GetWidth(),
				Height: rv.children[idx].GetHeight() + 1,
			})
			skip = 0
		} else {
			skip++
			if skip == len(rv.children) {
				break
			}
		}
		idx = (idx + 1) % len(rv.children)
	}
	return rv
}
