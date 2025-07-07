// Package grid implements functions to walk a 2D grid using various algorithms.
package grid

const (
	// PATTERN_INVALID represents an invalid pattern.
	PATTERN_INVALID = iota
	// PATTERN_SPIRAL represents a pattern that starts in the centre and grows counter clockwise.
	PATTERN_SPIRAL
	// PATTERN_LINEAR represents an iteration over every X and Y positions.
	PATTERN_LINEAR
)

const (
	directionUP = iota
	directionRIGHT
	directionDOWN
	directionLEFT
)

// GridPost represents a position in a 2D grid.
type GridPos struct {
	X int
	Y int
}

// WalkGrid returns a path that is the result of following the provided pattern.
func WalkGrid(sizeX int, sizeY int, pattern int) []GridPos {
	switch pattern {
	case PATTERN_SPIRAL:
		return walkGridSpiral(sizeX, sizeY)
	case PATTERN_LINEAR:
		return walkGridLinear(sizeX, sizeY)
	default:
		return []GridPos{}
	}
}

func walkGridLinear(sizeX int, sizeY int) []GridPos {
	path := []GridPos{}
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			path = append(path, GridPos{X: x, Y: y})
		}
	}

	return path
}

func walkGridSpiral(sizeX int, sizeY int) []GridPos {
	directions := []int{directionUP, directionRIGHT, directionDOWN, directionLEFT}
	path := []GridPos{}
	grid := make(map[GridPos]struct{})
	totalPositions := sizeX * sizeY
	directionIdx := 0

	// Start from the centre.
	cursor := GridPos{X: sizeX / 2, Y: sizeY / 2}
	path = append(path, cursor)
	grid[cursor] = struct{}{}
	walkedPositions := 1

	for {
		if walkedPositions == totalPositions {
			return path
		}

		dir := directions[directionIdx%len(directions)]
		switch dir {
		case directionUP:
			if _, ok := grid[GridPos{X: cursor.X, Y: cursor.Y - 1}]; ok {
				directionIdx--
				continue
			}
			cursor.Y--
			grid[cursor] = struct{}{}
			if insideGrid(cursor, sizeX, sizeY) {
				walkedPositions++
				path = append(path, cursor)
			}
			directionIdx++
			continue

		case directionDOWN:
			if _, ok := grid[GridPos{X: cursor.X, Y: cursor.Y + 1}]; ok {
				directionIdx--
				continue
			}
			cursor.Y++
			grid[cursor] = struct{}{}
			if insideGrid(cursor, sizeX, sizeY) {
				walkedPositions++
				path = append(path, cursor)
			}
			directionIdx++
			continue

		case directionLEFT:
			if _, ok := grid[GridPos{X: cursor.X - 1, Y: cursor.Y}]; ok {
				directionIdx--
				continue
			}
			cursor.X--
			grid[cursor] = struct{}{}
			if insideGrid(cursor, sizeX, sizeY) {
				walkedPositions++
				path = append(path, cursor)
			}
			directionIdx++
			continue

		case directionRIGHT:
			if _, ok := grid[GridPos{X: cursor.X + 1, Y: cursor.Y}]; ok {
				directionIdx--
				continue
			}
			cursor.X++
			grid[cursor] = struct{}{}
			if insideGrid(cursor, sizeX, sizeY) {
				walkedPositions++
				path = append(path, cursor)
			}
			directionIdx++
			continue
		}
	}
}

func insideGrid(cursor GridPos, sizeX, sizeY int) bool {
	return cursor.X >= 0 && cursor.X < sizeX && cursor.Y >= 0 && cursor.Y < sizeY
}
