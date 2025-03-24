package serde

import (
	"fmt"
	"io"
	"strings"

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
		// Try to extract the line number and value from the error
		if strings.Contains(err.Error(), "cannot unmarshal") {
			// Split the error message to get the line number
			parts := strings.Split(err.Error(), "line ")
			if len(parts) > 1 {
				lineInfo := strings.Split(parts[1], ":")
				if len(lineInfo) > 1 {
					lineNum := lineInfo[0]
					value := strings.TrimSpace(lineInfo[1])
					return nil, fmt.Errorf("invalid texture type at line %s: %s. Valid types are: constant, image, noise", lineNum, value)
				}
			}
		}
		return nil, err
	}

	return out, nil
}
