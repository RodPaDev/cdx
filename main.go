package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const (
	FILE_OBJECT_HEIGHT             = 5
	FILE_OBJECT_WIDTH              = 25
	FILE_OBJECT_HORIZONTAL_PADDING = 2
	FILE_OBJECT_VERTICAL_PADDING   = 2
	TOP_BAR_HEIGHT                 = 3
	BOTTOM_BAR_HEIGHT              = 3
	BORDER_SIZE                    = 1
)

var (
	borderColor   = lipgloss.Color("#2abbae")
	selectedColor = lipgloss.Color("#dadb83")

	styleScreen = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor)

	styleTile = lipgloss.NewStyle().
			Width(FILE_OBJECT_WIDTH).
			Height(FILE_OBJECT_HEIGHT).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor)

	styleTopBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false).
			BorderForeground(borderColor)

	styleBottomBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderLeft(false).
			BorderRight(false).
			BorderBottom(false).
			BorderForeground(borderColor)
)

type state struct {
	currentPath       string
	coordinateIdx     [2]int
	viewportRowOffset int
}

type model struct {
	width, height int
	state         state
	objects       []FileSystemObject
	rows, cols    int
}

func initModel(path string) model {
	return model{
		state: state{
			currentPath:   path,
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

		contentWidth := m.width - (2 * BORDER_SIZE)
		contentHeight := m.height - (2 * BORDER_SIZE)

		availableHeight := contentHeight - TOP_BAR_HEIGHT - BOTTOM_BAR_HEIGHT
		m.rows = availableHeight / (FILE_OBJECT_HEIGHT + FILE_OBJECT_VERTICAL_PADDING)
		m.cols = contentWidth / (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)

		if m.rows < 1 {
			m.rows = 1
		}
		if m.cols < 1 {
			m.cols = 1
		}

		m.objects = listObjects(m.state.currentPath)

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
	if idx >= len(m.objects) && len(m.objects) > 0 {
		last := len(m.objects) - 1
		m.state.coordinateIdx[0] = (last / m.cols) - m.state.viewportRowOffset
		if m.state.coordinateIdx[0] < 0 {
			m.state.coordinateIdx[0] = 0
			m.state.viewportRowOffset = last / m.cols
		}
		m.state.coordinateIdx[1] = last % m.cols
	}

	return m, nil
}

func (m model) View() string {
	contentWidth := m.width - (2 * BORDER_SIZE)
	contentHeight := m.height - (2 * BORDER_SIZE)

	explorerHeight := contentHeight - TOP_BAR_HEIGHT - BOTTOM_BAR_HEIGHT + 2

	topBar := styleTopBar.
		Width(contentWidth).
		Render(m.state.currentPathBreadcrumb(contentWidth))

	var fileExplorerRows []string

	// Create grid display for files and directories
	for rowIdx := 0; rowIdx < m.rows; rowIdx++ {
		var cols []string
		for colIdx := 0; colIdx < m.cols; colIdx++ {
			objectIdx := (m.state.viewportRowOffset+rowIdx)*m.cols + colIdx

			if objectIdx >= len(m.objects) {
				cols = append(cols, strings.Repeat(" ", FILE_OBJECT_WIDTH))
				continue
			}

			style := styleTile
			if rowIdx == m.state.coordinateIdx[0] && colIdx == m.state.coordinateIdx[1] {
				style = style.
					BorderForeground(selectedColor).
					Foreground(selectedColor)
			} else {
				style = style.BorderForeground(borderColor)
			}

			cols = append(cols, style.Render(
				renderFileTile(m.objects[objectIdx], FILE_OBJECT_WIDTH-2),
			))
		}
		fileExplorerRows = append(fileExplorerRows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	// Center the grid horizontally
	gridWidth := m.cols * (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)
	marginLeft := (contentWidth - gridWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}

	fileExplorer := lipgloss.NewStyle().
		MarginLeft(marginLeft).
		Height(explorerHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, fileExplorerRows...))

	navItems := []string{
		"h/j/k/l - move",
		"⏎ - open/navigate",
		"⌫ - up",
		"q - quit",
	}

	totalItemLength := 0
	for _, item := range navItems {
		totalItemLength += len(item)
	}
	spacesNeeded := contentWidth - totalItemLength - 2 // Account for padding
	spacesPerGap := 1
	if len(navItems) > 1 {
		spacesPerGap = spacesNeeded / (len(navItems) - 1)
		if spacesPerGap < 1 {
			spacesPerGap = 1
		}
	}

	navText := navItems[0]
	for i := 1; i < len(navItems); i++ {
		navText += strings.Repeat(" ", spacesPerGap) + navItems[i]
	}

	bottomBar := styleBottomBar.
		Width(contentWidth).
		Render(navText)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		fileExplorer,
		bottomBar,
	)

	return styleScreen.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent)
}

func getHomeDir() string {
	// First try HOME (Unix) or USERPROFILE (Windows)
	if runtime.GOOS == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home
		}
	} else {
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
	}

	usr, err := user.Current()
	if err == nil && usr.HomeDir != "" {
		return usr.HomeDir
	}

	return "/"
}

func main() {
	var argPath string
	if len(os.Args) > 1 {
		argPath = os.Args[1]
	} else {
		var err error
		argPath, err = os.Getwd()
		if err != nil {
			argPath = getHomeDir()
		}
	}
	fileInfo, err := os.Lstat(argPath)
	if err != nil {
		log.Fatalf("Could not get lstat of %s", argPath)
	}

	if !fileInfo.IsDir() {
		segments := strings.Split(argPath, "/")
		argPath = strings.Join(segments[0:len(segments)-1], "/")
	}

	p := tea.NewProgram(initModel(argPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
