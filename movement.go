package main

// MoveLeft moves the cursor one column to the left within the current viewport.
// If the cursor is already at the leftmost column (index 0), it wraps around to the last column.
func (s *state) MoveLeft(cols int) {
	if s.coordinateIdx[1] > 0 {
		// Normal move: decrement column index
		s.coordinateIdx[1] -= 1
	} else {
		// Wraparound: jump to last column index
		s.coordinateIdx[1] = cols - 1
	}
}

// MoveRight moves the cursor one column to the right within the current viewport.
// If the cursor is already at the last column, it wraps back to the first column (index 0).
func (s *state) MoveRight(cols int) {
	if s.coordinateIdx[1] < cols-1 {
		// Normal move: increment column index
		s.coordinateIdx[1] += 1
	} else {
		// Wraparound: reset column index to 0
		s.coordinateIdx[1] = 0
	}
}

// MoveDown moves the cursor one row down within the grid.
// If the cursor reaches the bottom visible row, it scrolls the viewport down instead.
// If already at the last row in the full grid, it wraps to the top.
func (s *state) MoveDown(rows, cols, objectsLen int) {
	// Compute current absolute row index (relative to full list, not just viewport)
	currentGlobalRow := s.viewportRowOffset + s.coordinateIdx[0]

	// Determine the last global row that contains an object
	lastGlobalRow := (objectsLen - 1) / cols

	if currentGlobalRow == lastGlobalRow {
		// If we're on the last row already, wrap to the very top
		s.viewportRowOffset = 0
		s.coordinateIdx = [2]int{0, 0}
		return
	}

	// Try to move one row down
	nextCoordRow := s.coordinateIdx[0] + 1
	nextGlobalIdx := (s.viewportRowOffset+nextCoordRow)*cols + s.coordinateIdx[1]

	if nextCoordRow < rows && nextGlobalIdx < objectsLen {
		// The next row fits within the viewport and contains a valid object
		s.coordinateIdx[0] = nextCoordRow
	} else {
		// The next row would be offscreen or empty, so scroll the viewport instead
		s.viewportRowOffset += 1
	}
}

// MoveUp moves the cursor one row up within the grid.
// If the cursor is already at the top of the viewport, it scrolls up.
// If already at the top of the list, it wraps to the last visible object.
func (s *state) MoveUp(rows, cols, objectsLen int) {
	if s.coordinateIdx[0] > 0 {
		// If not at top row of viewport, move up locally
		s.coordinateIdx[0] -= 1
		return
	}

	if s.viewportRowOffset > 0 {
		// At top of viewport, but there's more content above: scroll up
		s.viewportRowOffset -= 1
		return
	}

	// We're already at the top of the list. Wrap to the last visible row of the last page.
	lastIdx := objectsLen - 1         // Index of final item
	totalRows := (lastIdx / cols) + 1 // Total number of full + partial rows

	// Compute how much to scroll so the last row is visible
	s.viewportRowOffset = totalRows - rows
	if s.viewportRowOffset < 0 {
		// If file list is shorter than screen, no need to offset
		s.viewportRowOffset = 0
	}

	// Set cursor to the final item on screen
	s.coordinateIdx[0] = (lastIdx / cols) - s.viewportRowOffset // Final row index, adjusted to viewport
	s.coordinateIdx[1] = lastIdx % cols                         // Final column index
}
