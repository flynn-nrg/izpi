package render

const (
	PATTERN_INVALID = iota
	PATTERN_SPIRAL
	PATTERN_LINEAR
)

const (
	DIRECTION_UP = iota
	DIRECTION_RIGHT
	DIRECTION_DOWN
	DIRECTION_LEFT
)

type gridPos struct {
	X int
	Y int
}

func walkGrid(sizeX int, sizeY int, pattern int) []gridPos {
	switch pattern {
	case PATTERN_SPIRAL:
		return walkGridSpiral(sizeX, sizeY)
	case PATTERN_LINEAR:
		return walkGridLinear(sizeX, sizeY)
	default:
		return []gridPos{}
	}
}

func walkGridLinear(sizeX int, sizeY int) []gridPos {
	path := []gridPos{}
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			path = append(path, gridPos{X: x, Y: y})
		}
	}

	return path
}

func walkGridSpiral(sizeX int, sizeY int) []gridPos {
	directions := []int{DIRECTION_UP, DIRECTION_RIGHT, DIRECTION_DOWN, DIRECTION_LEFT}
	path := []gridPos{}
	grid := make(map[gridPos]struct{})
	totalPositions := sizeX * sizeY
	walkedPositions := 1
	directionIdx := 0

	// Start from the centre.
	cursor := gridPos{X: sizeX / 2, Y: sizeY / 2}
	path = append(path, cursor)
	grid[cursor] = struct{}{}

	for {
		if walkedPositions == totalPositions {
			return path
		}

		dir := directions[directionIdx%len(directions)]
		switch dir {
		case DIRECTION_UP:
			if _, ok := grid[gridPos{X: cursor.X, Y: cursor.Y - 1}]; ok {
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

		case DIRECTION_DOWN:
			if _, ok := grid[gridPos{X: cursor.X, Y: cursor.Y + 1}]; ok {
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

		case DIRECTION_LEFT:
			if _, ok := grid[gridPos{X: cursor.X - 1, Y: cursor.Y}]; ok {
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

		case DIRECTION_RIGHT:
			if _, ok := grid[gridPos{X: cursor.X + 1, Y: cursor.Y}]; ok {
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

func insideGrid(cursor gridPos, sizeX, sizeY int) bool {
	return cursor.X >= 0 && cursor.X < sizeX && cursor.Y >= 0 && cursor.Y < sizeY
}
