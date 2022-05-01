package postprocess

import (
	"image"

	"gitlab.com/flynn-nrg/izpi/pkg/scene"
)

// Pipeline represents a filter pipeline.
type Pipeline struct {
	filters []Filter
}

// NewPipeline returns a new filter pipeline.
func NewPipeline(filters []Filter) *Pipeline {
	return &Pipeline{
		filters: filters,
	}
}

// Apply applies all the filters in the pipeline in the order they were added to the slice.
func (p *Pipeline) Apply(i image.Image, scene *scene.Scene) error {
	for _, f := range p.filters {
		err := f.Apply(i, scene)
		if err != nil {
			return err
		}

	}

	return nil
}
