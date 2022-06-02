package serde

import (
	"io"

	"gopkg.in/yaml.v3"
)

var _ Serde = (*Yaml)(nil)

// Yaml represents a struct representation of a scene that is seralised as YAML data.
type Yaml struct {
}

func (yml *Yaml) Serialise(scene *Scene, w io.Writer) error {
	out, err := yaml.Marshal(scene)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (yml *Yaml) Deserialise(r io.Reader) (*Scene, error) {
	out := &Scene{}
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(in, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
