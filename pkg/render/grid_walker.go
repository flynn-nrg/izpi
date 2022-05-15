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
	unsuccessfulTries := 0

	for {
		if walkedPositions == totalPositions || unsuccessfulTries > 4 {
			for y := sizeY - 1; y >= 0; y-- {
				for x := 0; x < sizeX; x++ {
					pos := gridPos{X: x, Y: y}
					if _, ok := grid[pos]; !ok {
						path = append(path, pos)
					}
				}
			}
			return path
		}

		dir := directions[directionIdx%len(directions)]
		switch dir {
		case DIRECTION_UP:
			if cursor.Y == 0 {
				directionIdx++
				unsuccessfulTries++
				continue
			}
			if _, ok := grid[gridPos{X: cursor.X, Y: cursor.Y - 1}]; ok {
				unsuccessfulTries++
				directionIdx--
				continue
			}
			cursor.Y--
			grid[cursor] = struct{}{}
			walkedPositions++
			path = append(path, cursor)
			directionIdx++
			unsuccessfulTries = 0
			continue

		case DIRECTION_DOWN:
			if cursor.Y+1 == sizeY {
				directionIdx++
				continue
			}
			if _, ok := grid[gridPos{X: cursor.X, Y: cursor.Y + 1}]; ok {
				unsuccessfulTries++
				directionIdx--
				continue
			}
			cursor.Y++
			grid[cursor] = struct{}{}
			walkedPositions++
			path = append(path, cursor)
			directionIdx++
			unsuccessfulTries = 0
			continue

		case DIRECTION_LEFT:
			if cursor.X == 0 {
				directionIdx++
				unsuccessfulTries++
				continue
			}
			if _, ok := grid[gridPos{X: cursor.X - 1, Y: cursor.Y}]; ok {
				unsuccessfulTries++
				directionIdx--
				continue
			}
			cursor.X--
			grid[cursor] = struct{}{}
			walkedPositions++
			path = append(path, cursor)
			directionIdx++
			unsuccessfulTries = 0
			continue

		case DIRECTION_RIGHT:
			if cursor.X+1 == sizeX {
				directionIdx++
				unsuccessfulTries++
				continue
			}
			if _, ok := grid[gridPos{X: cursor.X + 1, Y: cursor.Y}]; ok {
				unsuccessfulTries++
				directionIdx--
				continue
			}
			cursor.X++
			grid[cursor] = struct{}{}
			walkedPositions++
			path = append(path, cursor)
			directionIdx++
			unsuccessfulTries = 0
			continue
		}
	}
}
