package responsive

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCase1(t *testing.T) {
	root := struct {
		ResponsiveView
	}{}
	child1 := struct {
		ResponsiveView
	}{
		ResponsiveView{minHeight: 5, minWidth: 5, maxHeight: 20, maxWidth: 20},
	}
	child2 := struct {
		ResponsiveView
	}{
		ResponsiveView{minHeight: 5, minWidth: 5, maxHeight: 20, maxWidth: 20},
	}

	root.children = append(root.children, &child1)
	root.children = append(root.children, &child2)

	root.SetSize(tea.WindowSizeMsg{
		Width:  20,
		Height: 20,
	})

	assert.Equal(t, 10, child1.width)
	assert.Equal(t, 10, child1.height)

	assert.Equal(t, 10, child2.width)
	assert.Equal(t, 10, child2.height)

	assert.Equal(t, 20, root.width)
	assert.Equal(t, 20, root.height)
}

func TestCase2(t *testing.T) {
	root := struct {
		ResponsiveView
	}{}
	child1 := struct {
		ResponsiveView
	}{}
	child2 := struct {
		ResponsiveView
	}{}

	root.children = append(root.children, &child1)
	root.children = append(root.children, &child2)

	root.SetSize(tea.WindowSizeMsg{
		Width:  20,
		Height: 20,
	})

	assert.Equal(t, 10, child1.width)
	assert.Equal(t, 10, child1.height)

	assert.Equal(t, 10, child2.width)
	assert.Equal(t, 10, child2.height)

	assert.Equal(t, 20, root.width)
	assert.Equal(t, 20, root.height)
}

func TestCase3(t *testing.T) {
	root := struct {
		ResponsiveView
	}{}
	child1 := struct {
		ResponsiveView
	}{ResponsiveView{minHeight: 0, minWidth: 0, maxHeight: 20, maxWidth: 20}}
	child2 := struct {
		ResponsiveView
	}{ResponsiveView{minHeight: 5, minWidth: 5, maxHeight: 20, maxWidth: 20}}

	root.children = append(root.children, &child1)
	root.children = append(root.children, &child2)

	root.SetSize(tea.WindowSizeMsg{
		Width:  20,
		Height: 20,
	})

	assert.Equal(t, 8, child1.width)
	assert.Equal(t, 8, child1.height)

	assert.Equal(t, 12, child2.width)
	assert.Equal(t, 12, child2.height)

	assert.Equal(t, 20, root.width)
	assert.Equal(t, 20, root.height)
}
