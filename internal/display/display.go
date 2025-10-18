package display

type DisplayTile struct {
	Width  int
	Height int
	PosX   int
	PosY   int
	Pixels []float32
}

type Display interface {
	Start()
	Wait()
}
