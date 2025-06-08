package material

import "github.com/flynn-nrg/izpi/internal/texture"

type nonPBR struct{}

func (np *nonPBR) NormalMap() texture.Texture {
	return nil
}
