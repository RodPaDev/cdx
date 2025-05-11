package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatSize converts a file size in bytes to a human-readable string.
// Uses binary units: 1024 bytes = 1 KB, 1024 KB = 1 MB, etc.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		// No conversion needed, show bytes directly
		return fmt.Sprintf("%d B", bytes)
	}

	// Find the largest unit to use (KB, MB, GB, etc.) without going over
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp += 1
	}

	// Format with one decimal of precision, suffix unit with capital letter
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// renderFileTile returns a vertical, multi-line string that represents one file or directory.
// This includes name, modified date, size (or "-"), and vertical padding for layout balance.
func renderFileTile(obj FileSystemObject, width int) string {
	date := obj.ModTime.Format("2006-01-02") // Fixed format for consistency

	// Files show human-readable size; directories use "-"
	size := "-"
	if !obj.IsDir {
		size = formatSize(obj.Size)
	}

	// Align date and size at opposite ends of the line
	infoLine := spaceBetween([]string{date, size}, width)

	// Label as [D] for directory, [F] for file
	namePrefix := "F"
	if obj.IsDir {
		namePrefix = "D"
	}

	// Combine prefix and name; truncate with ellipsis if it doesn't fit
	name := truncateCenter(fmt.Sprintf("[%s] %s", namePrefix, obj.Name), width)

	// Final tile: top/bottom padding, name, spacer, and info
	return lipgloss.JoinVertical(lipgloss.Top,
		"",       // Top spacer
		name,     // Name line
		"",       // Mid spacer
		infoLine, // Info (date + size)
		"",       // Bottom spacer
	)
}

// truncateCenter shortens a string by replacing the center with an ellipsis (…)
// Ensures the string does not exceed the given visual width.
func truncateCenter(s string, width int) string {
	if lipgloss.Width(s) <= width {
		// No need to truncate
		return s
	}

	// Remove characters from the center to make space for ellipsis
	runes := []rune(s)
	cut := (width - 1) / 2
	return string(runes[:cut]) + "…" + string(runes[len(runes)-cut:])
}

// openCurrentPath resets the grid view when entering a new directory.
// It reinitializes the cursor, scroll offset, and reloads the file list.
func (m *model) openCurrentPath() {
	if m.state.currentPath == "" {
		m.state.currentPath = "/" // Normalize empty path as root
	}

	// Reset viewport and cursor position
	m.state.coordinateIdx = [2]int{0, 0}
	m.state.viewportRowOffset = 0

	// Fetch directory contents
	m.objects = listObjects(m.state.currentPath)
}

// handleSelection determines the action when the user presses Enter:
// If the item is a directory, enter it; if it's a file, open it using the OS.
func (m *model) handleSelection() {
	idx := (m.state.viewportRowOffset+m.state.coordinateIdx[0])*m.cols + m.state.coordinateIdx[1]

	if idx >= len(m.objects) {
		return // Invalid index (likely empty space), do nothing
	}

	obj := m.objects[idx]
	if obj.IsDir {
		// Change into directory and refresh view
		m.state.currentPath = obj.Path
		m.openCurrentPath()
	} else {
		// Open the file in the system default app
		OpenFile(obj.Path)
	}
}

// currentPathBreadcrumb builds a path display for the top bar (e.g., /usr/bin/go).
// If the full path is too wide, truncates the middle with an ellipsis to preserve start/end.
func (s state) currentPathBreadcrumb(maxWidth int) string {
	if s.currentPath == "/" {
		return " /" // Special case: root only
	}

	// Split path into parts and add leading slash to each
	parts := strings.Split(strings.TrimPrefix(s.currentPath, "/"), "/")
	segments := make([]string, len(parts))
	for i, part := range parts {
		segments[i] = " / " + part
	}

	full := strings.Join(segments, "")
	if lipgloss.Width(full) <= maxWidth {
		return full // Fits within screen
	}

	// If too wide, insert ellipsis in the middle
	ellipsis := " / ..."
	avail := maxWidth - lipgloss.Width(ellipsis)
	if avail <= 0 {
		return ellipsis // Not enough space for anything else
	}

	left := []string{}
	right := []string{}
	leftWidth := 0
	rightWidth := 0
	i := 0
	j := len(segments) - 1

	// Alternate adding from left and right until the space is full
	for i <= j {
		if leftWidth <= rightWidth {
			segW := lipgloss.Width(segments[i])
			if leftWidth+rightWidth+segW > avail {
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
			j--
		}
	}

	// Join left side + ellipsis + right side
	return strings.Join(left, "") + ellipsis + strings.Join(right, "")
}

// spaceBetween arranges strings evenly within a given width by adding spaces between them.
// It guarantees at least one space between items and distributes remaining space as evenly as possible.
func spaceBetween(items []string, totalWidth int) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}

	totalItemsWidth := 0
	for _, item := range items {
		totalItemsWidth += lipgloss.Width(item)
	}

	gaps := len(items) - 1
	spaceRemaining := totalWidth - totalItemsWidth
	if spaceRemaining < gaps {
		spaceRemaining = gaps // Minimum one space per gap
	}
	spacePerGap := spaceRemaining / gaps
	extraSpace := spaceRemaining % gaps

	var builder strings.Builder
	for i, item := range items {
		builder.WriteString(item)
		if i < gaps {
			spaces := spacePerGap
			if i < extraSpace {
				spaces += 1 // Add one extra space to first 'extraSpace' gaps
			}
			builder.WriteString(strings.Repeat(" ", spaces))
		}
	}
	return builder.String()
}
