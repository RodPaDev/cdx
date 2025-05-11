// movement.go
package main

// MoveLeft moves the focus one column to the left, wrapping around.
func (s *state) MoveLeft(cols int) {
	if s.coordinateIdx[1] > 0 {
		s.coordinateIdx[1]--
	} else {
		s.coordinateIdx[1] = cols - 1
	}
}

// MoveRight moves the focus one column to the right, wrapping around.
func (s *state) MoveRight(cols int) {
	if s.coordinateIdx[1] < cols-1 {
		s.coordinateIdx[1]++
	} else {
		s.coordinateIdx[1] = 0
	}
}

// MoveDown moves the focus down one row, or scrolls the viewport if at the bottom.
func (s *state) MoveDown(rows, cols, objectsLen int) {
	currentGlobalRow := s.viewportRowOffset + s.coordinateIdx[0]
	lastGlobalRow := (objectsLen - 1) / cols

	if currentGlobalRow == lastGlobalRow {
		// wrap to top
		s.viewportRowOffset = 0
		s.coordinateIdx = [2]int{0, 0}
		return
	}

	nextCoordRow := s.coordinateIdx[0] + 1
	nextGlobalIdx := (s.viewportRowOffset+nextCoordRow)*cols + s.coordinateIdx[1]

	if nextCoordRow < rows && nextGlobalIdx < objectsLen {
		s.coordinateIdx[0] = nextCoordRow
	} else {
		s.viewportRowOffset++
	}
}

// MoveUp moves the focus up one row, or scrolls the viewport if at the top.
func (s *state) MoveUp(rows, cols, objectsLen int) {
	if s.coordinateIdx[0] > 0 {
		s.coordinateIdx[0]--
		return
	}
	if s.viewportRowOffset > 0 {
		s.viewportRowOffset--
		return
	}

	// wrap to bottom
	lastIdx := objectsLen - 1
	totalRows := (lastIdx / cols) + 1
	s.viewportRowOffset = totalRows - rows
	if s.viewportRowOffset < 0 {
		s.viewportRowOffset = 0
	}
	s.coordinateIdx[0] = (lastIdx / cols) - s.viewportRowOffset
	s.coordinateIdx[1] = lastIdx % cols
}
