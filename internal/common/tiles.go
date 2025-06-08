package common

var stepSizes = []int{32, 25, 24, 20, 16, 12, 10, 8, 5, 4}

// Tiles computes a tile size for the provided width and height.
func Tiles(sizeX, sizeY int) (int, int) {
	var stepSizeX, stepSizeY int

	for _, size := range stepSizes {
		if sizeX%size == 0 {
			stepSizeX = size
			break
		}
	}

	for _, size := range stepSizes {
		if sizeY%size == 0 {
			stepSizeY = size
			break
		}
	}

	return stepSizeX, stepSizeY
}
