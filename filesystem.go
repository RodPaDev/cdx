package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type FileSystemObject struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

func ListObjects(path string) []FileSystemObject {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var objects []FileSystemObject

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Fatal("Cannot get entry info", err)
		}

		entryPath := filepath.Join(path, entry.Name())
		if absPath, err := filepath.Abs(entryPath); err == nil {
			entryPath = absPath
		} else {
			log.Fatal("Cannot get entryPath abs", err)
		}

		objects = append(objects, FileSystemObject{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Path:    entryPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return objects
}

func (f FileSystemObject) ShortName(maxWidth int) string {
	if maxWidth == 0 {
		maxWidth = 10
	}

	name := f.Name
	if lipgloss.Width(name) <= maxWidth {
		return name
	}

	dots := "..."
	dotsWidth := lipgloss.Width(dots)
	allowedWidth := maxWidth - dotsWidth

	currentWidth := 0
	cutoffIndex := 0
	for i, r := range name {
		w := lipgloss.Width(string(r))
		if currentWidth+w > allowedWidth {
			break
		}
		currentWidth += w
		cutoffIndex = i + 1
	}

	return name[:cutoffIndex] + dots
}

func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default: // Linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
