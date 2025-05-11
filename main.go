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

// Layout constants defining the structure and spacing of the UI
const (
	FILE_OBJECT_HEIGHT             = 5  // Number of terminal rows each file/directory tile occupies
	FILE_OBJECT_WIDTH              = 25 // Number of terminal columns each tile occupies
	FILE_OBJECT_HORIZONTAL_PADDING = 2  // Horizontal padding between tiles (in columns)
	FILE_OBJECT_VERTICAL_PADDING   = 2  // Vertical padding between tiles (in rows)
	TOP_BAR_HEIGHT                 = 3  // Height of the top bar showing path (breadcrumb)
	BOTTOM_BAR_HEIGHT              = 3  // Height of the bottom bar showing key hints
	BORDER_SIZE                    = 1  // Thickness of the screen border (applied on all sides)
)

// Color and style definitions for various UI components
var (
	borderColor   = lipgloss.Color("#2abbae") // Default border color (teal)
	selectedColor = lipgloss.Color("#dadb83") // Highlight color (yellow-like)

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

// state contains all mutable information regarding navigation and viewport
type state struct {
	currentPath       string // Current path shown on screen
	coordinateIdx     [2]int // Cursor grid position: [row, col] within the visible area
	viewportRowOffset int    // Vertical offset from the top of the file list (for scrolling)
}

// model represents the application state, UI layout, and data
type model struct {
	width, height int                // Dimensions of the terminal window (in characters)
	state         state              // Navigation state
	objects       []FileSystemObject // Flat list of all objects (files and dirs) in current directory
	rows, cols    int                // Grid size: number of visible rows and columns based on screen size
}

// initModel returns a fresh model for a given path with initial position at top-left
func initModel(path string) model {
	return model{
		state: state{
			currentPath:   path,
			coordinateIdx: [2]int{0, 0}, // Start selection at the top-left tile
		},
	}
}

// Init is called by the Bubble Tea framework to initialize terminal mode and screen size
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tea.ClearScreen, tea.WindowSize())
}

// Update handles terminal events like key presses and window resizes
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Save new terminal size
		m.width = msg.Width
		m.height = msg.Height

		// Compute usable width/height after removing outer borders
		contentWidth := m.width - (2 * BORDER_SIZE)
		contentHeight := m.height - (2 * BORDER_SIZE)

		// Deduct top and bottom bar height from total usable height
		availableHeight := contentHeight - TOP_BAR_HEIGHT - BOTTOM_BAR_HEIGHT

		// Determine how many full tiles (including spacing) fit vertically
		m.rows = availableHeight / (FILE_OBJECT_HEIGHT + FILE_OBJECT_VERTICAL_PADDING)
		// Determine how many full tiles (including spacing) fit horizontally
		m.cols = contentWidth / (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)

		// Ensure there’s always at least 1 row and 1 column to prevent divide-by-zero or invisible UI
		if m.rows < 1 {
			m.rows = 1
		}
		if m.cols < 1 {
			m.cols = 1
		}

		// Reload file list for new screen layout
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
			// Move to parent directory by trimming last path segment
			segments := strings.Split(m.state.currentPath, "/")
			m.state.currentPath = strings.Join(segments[:len(segments)-1], "/")
			m.openCurrentPath()
		}
	}

	// Ensure cursor remains within bounds of updated object list
	idx := (m.state.viewportRowOffset+m.state.coordinateIdx[0])*m.cols + m.state.coordinateIdx[1]

	if idx >= len(m.objects) && len(m.objects) > 0 {
		last := len(m.objects) - 1

		// Calculate object's row and column in the full list (0-based)
		lastRow := last / m.cols // Full list row index
		lastCol := last % m.cols // Column index within that row
		visibleRow := lastRow - m.state.viewportRowOffset

		// If last row is above current viewport, snap viewport to show it
		if visibleRow < 0 {
			m.state.viewportRowOffset = lastRow
			visibleRow = 0
		}

		// Update cursor position to the last object in the list
		m.state.coordinateIdx[0] = visibleRow
		m.state.coordinateIdx[1] = lastCol
	}

	return m, nil
}

// View constructs the entire screen output as a string and returns it.
// It builds the top bar (path), file grid, and bottom bar (key hints),
// and arranges them vertically within the available content area.
func (m model) View() string {
	// Calculate dimensions of usable area inside the border
	contentWidth := m.width - (2 * BORDER_SIZE)
	contentHeight := m.height - (2 * BORDER_SIZE)

	// Calculate the height left for the file explorer section (grid of files)
	// Add 2 to prevent clipping due to border interactions or rounding
	explorerHeight := contentHeight - TOP_BAR_HEIGHT - BOTTOM_BAR_HEIGHT + 2

	// Render the top bar: breadcrumb-style path navigation
	topBar := styleTopBar.
		Width(contentWidth).
		Render(m.state.currentPathBreadcrumb(contentWidth))

	var fileExplorerRows []string

	// Build the 2D grid row by row
	for rowIdx := 0; rowIdx < m.rows; rowIdx++ {
		var cols []string

		for colIdx := 0; colIdx < m.cols; colIdx++ {
			// Translate 2D grid coords into 1D index in m.objects
			objectIdx := (m.state.viewportRowOffset+rowIdx)*m.cols + colIdx

			if objectIdx >= len(m.objects) {
				// If the grid cell is out-of-bounds (no object), fill it with blank space
				cols = append(cols, strings.Repeat(" ", FILE_OBJECT_WIDTH))
				continue
			}

			// Highlight tile if it's currently selected
			style := styleTile
			if rowIdx == m.state.coordinateIdx[0] && colIdx == m.state.coordinateIdx[1] {
				style = style.
					BorderForeground(selectedColor).
					Foreground(selectedColor)
			}

			// Render a single tile (file or folder)
			cols = append(cols, style.Render(
				renderFileTile(m.objects[objectIdx], FILE_OBJECT_WIDTH-2),
			))
		}

		// Concatenate all tiles horizontally to form a visual row
		fileExplorerRows = append(fileExplorerRows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	// Center the entire grid horizontally within the content area
	gridWidth := m.cols * (FILE_OBJECT_WIDTH + FILE_OBJECT_HORIZONTAL_PADDING)
	marginLeft := (contentWidth - gridWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}

	// Final rendering of the file explorer block
	fileExplorer := lipgloss.NewStyle().
		MarginLeft(marginLeft).
		Height(explorerHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, fileExplorerRows...))

	// Key hint items shown at bottom
	navItems := []string{
		"h/j/k/l - move",
		"⏎ - open/navigate",
		"⌫ - up",
		"q - quit",
	}

	// Calculate total fixed length of nav items (text only)
	totalItemLength := 0
	for _, item := range navItems {
		totalItemLength += len(item)
	}
	// Calculate spacing between nav items so they spread across the width
	spacesNeeded := contentWidth - totalItemLength - 2 // Allow for some margin
	spacesPerGap := 1
	if len(navItems) > 1 {
		spacesPerGap = spacesNeeded / (len(navItems) - 1)
		if spacesPerGap < 1 {
			spacesPerGap = 1
		}
	}

	// Build nav hint line with spacing between items
	navText := navItems[0]
	for i := 1; i < len(navItems); i++ {
		navText += strings.Repeat(" ", spacesPerGap) + navItems[i]
	}

	// Render the bottom bar with navigation info
	bottomBar := styleBottomBar.
		Width(contentWidth).
		Render(navText)

	// Combine top bar, file grid, and bottom bar vertically
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		fileExplorer,
		bottomBar,
	)

	// Wrap the entire UI inside a border and return final render string
	return styleScreen.
		Width(contentWidth).
		Height(contentHeight).
		Render(mainContent)
}

// getHomeDir attempts to find the user's home directory in a cross-platform way.
// Prioritizes environment variables, then OS user info, then defaults to root.
func getHomeDir() string {
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

	// Fallback if everything fails
	return "/"
}

// main is the entry point of the application.
// It determines the initial path to explore and starts the Bubble Tea program.
func main() {
	var argPath string

	if len(os.Args) > 1 {
		// Use user-supplied argument if provided
		argPath = os.Args[1]
	} else {
		// Attempt to use current working directory
		var err error
		argPath, err = os.Getwd()
		if err != nil {
			// Fall back to home directory if CWD fails
			argPath = getHomeDir()
		}
	}

	// Validate the path exists
	fileInfo, err := os.Lstat(argPath)
	if err != nil {
		log.Fatalf("Could not get lstat of %s", argPath)
	}

	// If user gave a file instead of a directory, use the parent folder instead
	if !fileInfo.IsDir() {
		segments := strings.Split(argPath, "/")
		argPath = strings.Join(segments[0:len(segments)-1], "/")
	}

	// Start the terminal UI program using Bubble Tea
	p := tea.NewProgram(initModel(argPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
