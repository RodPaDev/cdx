package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func renderFileTile(obj FileSystemObject, width int) string {
	date := obj.ModTime.Format("2006-01-02")
	size := "-"
	if !obj.IsDir {
		size = formatSize(obj.Size)
	}
	infoLine := spaceBetween([]string{date, size}, width)

	namePrefix := "F"
	if obj.IsDir {
		namePrefix = "D"
	}

	name := truncateCenter(fmt.Sprintf("[%s] %s", namePrefix, obj.Name), width)

	return lipgloss.JoinVertical(lipgloss.Top,
		"",
		name,
		"",
		infoLine,
		"",
	)
}

func truncateCenter(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	cut := (width - 1) / 2
	return string(runes[:cut]) + "…" + string(runes[len(runes)-cut:])
}

func (m *model) openCurrentPath() {
	if m.state.currentPath == "" {
		m.state.currentPath = "/"
	}
	m.state.coordinateIdx = [2]int{0, 0}
	m.state.viewportRowOffset = 0
	m.objects = listObjects(m.state.currentPath)
}

func (m *model) handleSelection() {
	idx := (m.state.viewportRowOffset+m.state.coordinateIdx[0])*m.cols + m.state.coordinateIdx[1]
	if idx >= len(m.objects) {
		return
	}
	if m.objects[idx].IsDir {
		m.state.currentPath = m.objects[idx].Path
		m.openCurrentPath()
	} else {
		OpenFile(m.objects[idx].Path)
	}
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
		spaceRemaining = gaps
	}
	spacePerGap := spaceRemaining / gaps
	extraSpace := spaceRemaining % gaps

	var builder strings.Builder
	for i, item := range items {
		builder.WriteString(item)
		if i < gaps {
			spaces := spacePerGap
			if i < extraSpace {
				spaces++
			}
			builder.WriteString(strings.Repeat(" ", spaces))
		}
	}
	return builder.String()
}
