package material

import "github.com/flynn-nrg/izpi/pkg/texture"

type nonPBR struct{}

func (np *nonPBR) NormalMap() texture.Texture {
	return nil
}
