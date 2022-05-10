package material

import "gitlab.com/flynn-nrg/izpi/pkg/texture"

type nonPBR struct{}

func (np *nonPBR) NormalMap() texture.Texture {
	return nil
}
