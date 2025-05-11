package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const (
	OUTSIDE_PADDING                = 2
	FILE_OBJECT_HEIGHT             = 5
	FILE_OBJECT_WIDTH              = 25
	FILE_OBJECT_HORIZONTAL_PADDING = 2
	FILE_OBJECT_VERTICAL_PADDING   = 2
)

var (
	styleTile = lipgloss.NewStyle().
			Width(FILE_OBJECT_WIDTH).
			Height(FILE_OBJECT_HEIGHT).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#2abbae"))

	styleTileSelected = styleTile.
				BorderBackground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#FAFAFA"))

	styleScreen = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#2abbae"))
)

type state struct {
	currentPath       string
	coordinateIdx     [2]int
	viewportRowOffset int
}

func (s state) currentPathBreadcrumb(maxWidth int) string {
	if s.currentPath == "/" {
		return " /"
	}

	parts := strings.Split(strings.TrimPrefix(s.currentPath, "/"), "/")
	segments := make([]string, len(parts))
	for i, part := range parts {
		segments[i] = " / " + part
	}

	full := strings.Join(segments, "")
	if lipgloss.Width(full) <= maxWidth {
		return full
	}

	ellipsis := " / ..."
	avail := maxWidth - lipgloss.Width(ellipsis)
	if avail <= 0 {
		return ellipsis
	}

	left := []string{}
	right := []string{}

	leftWidth := 0
	rightWidth := 0
	i := 0
	j := len(segments) - 1

	for i <= j {
		if leftWidth <= rightWidth {
			segW := lipgloss.Width(segments[i])
			if leftWidth+segW+rightWidth > avail {
				break
			}
			left = append(left, segments[i])
			leftWidth += segW
			i += 1
		} else {
			segW := lipgloss.Width(segments[j])
			if leftWidth+rightWidth+segW > avail {
				break
			}
			right = append([]string{segments[j]}, right...)
			rightWidth += segW
			j -= 1
		}
	}

	return strings.Join(left, "") + ellipsis + strings.Join(right, "")
}

type model struct {
	width, height int
	state         state
	objects       []FileSystemObject
	rows, cols    int
}

func (m *model) openCurrentPath() {
	if m.state.currentPath == "" {
		m.state.currentPath = "/"
	}
	m.state.coordinateIdx = [2]int{0, 0}
	m.state.viewportRowOffset = 0
	m.objects = ListObjects(m.state.currentPath)
}

func (m *model) handleSelection() {
	idx := (m.state.viewportRowOffset+m.state.coordinateIdx[0])*m.cols + m.state.coordinateIdx[1]
	if m.objects[idx].IsDir {
		m.state.currentPath = m.objects[idx].Path
		m.openCurrentPath()
	} else {
		OpenFile(m.objects[idx].Path)
	}
}

func initModel() model {
	return model{
		state: state{
			currentPath:   "/Users/patrickrodrigues",
			coordinateIdx: [2]int{0, 0},
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tea.ClearScreen, tea.WindowSize())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		usableH := m.height - OUTSIDE_PADDING
		usableW := m.width - OUTSIDE_PADDING
		m.rows = usableH / (FILE_OBJECT_HEIGHT + FILE_OBJECT_VERTICAL_PADDING)
		m.cols = usableW / (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)

		m.objects = ListObjects(m.state.currentPath)

	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			m.state.MoveLeft(m.cols)
		case "j":
			m.state.MoveDown(m.rows, m.cols, len(m.objects))
		case "k":
			m.state.MoveUp(m.rows, m.cols, len(m.objects))
		case "l":
			m.state.MoveRight(m.cols)
		case "q":
			return m, tea.Quit
		case tea.KeyEnter.String():
			m.handleSelection()
		case tea.KeyBackspace.String():
			segments := strings.Split(m.state.currentPath, "/")
			m.state.currentPath = strings.Join(strings.Split(m.state.currentPath, "/")[0:len(segments)-1], "/")
			m.openCurrentPath()
		}
	}

	idx := (m.state.viewportRowOffset+m.state.coordinateIdx[0])*m.cols + m.state.coordinateIdx[1]
	if idx >= len(m.objects) {
		last := len(m.objects) - 1
		m.state.coordinateIdx[0] = last/m.cols - m.state.viewportRowOffset
		m.state.coordinateIdx[1] = last % m.cols
	}

	return m, nil
}

func (m model) View() string {
	usableH := m.height - OUTSIDE_PADDING
	usableW := m.width - OUTSIDE_PADDING

	var rows []string
	for rowIdx := range m.rows {
		var cols []string
		for colIdx := range m.cols {
			objectIdx := (m.state.viewportRowOffset+rowIdx)*m.cols + colIdx

			if objectIdx >= len(m.objects) {
				cols = append(cols, "")
				continue
			}

			style := styleTile
			if rowIdx == m.state.coordinateIdx[0] && colIdx == m.state.coordinateIdx[1] {
				style = styleTileSelected
			}

			cols = append(cols, style.Render(
				lipgloss.JoinVertical(lipgloss.Center, "", m.objects[objectIdx].ShortName(FILE_OBJECT_WIDTH-1)),
			))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	gridW := m.cols * (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)
	grid := lipgloss.NewStyle().
		MarginLeft((usableW - gridW) / 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	return styleScreen.
		Width(usableW).
		Height(usableH).
		Render(lipgloss.JoinVertical(lipgloss.Top, m.state.currentPathBreadcrumb(m.width), grid))
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
