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

// FileSystemObject holds metadata for a file or directory
type FileSystemObject struct {
	Name    string    // Entry name
	Path    string    // Absolute path
	IsDir   bool      // True if directory
	Size    int64     // Size in bytes (files only)
	ModTime time.Time // Last modified time
}

// listObjects reads a directory and returns its entries as FileSystemObjects
func listObjects(path string) []FileSystemObject {
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
		absPath, err := filepath.Abs(entryPath)
		if err != nil {
			log.Fatal("Cannot get abs path", err)
		}

		objects = append(objects, FileSystemObject{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Path:    absPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return objects
}

// ShortName trims long filenames to fit within a tile width
func (f FileSystemObject) ShortName(maxWidth int) string {
	if maxWidth == 0 {
		maxWidth = 10
	}

	if lipgloss.Width(f.Name) <= maxWidth {
		return f.Name
	}

	// Truncate and add "..."
	dots := "..."
	allowed := maxWidth - lipgloss.Width(dots)

	width := 0
	// iterate over the string and add up the width of each character
	// until the total width exceeds the allowed width
	i := 0
	for ; i < len(f.Name); i++ {
		w := lipgloss.Width(string(f.Name[i]))
		if width+w > allowed {
			break
		}
		width += w
	}

	return f.Name[:i] + dots
}

// OpenFile opens a file using the OS default app
func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
