package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// nonSpectral provides stub implementations for spectral methods
// This allows existing materials to satisfy the Material interface
// without implementing spectral rendering capabilities
type nonSpectral struct{}

// SpectralScatter provides a stub implementation that returns false
// indicating that this material doesn't support spectral scattering
func (ns *nonSpectral) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	return nil, nil, false
}

// SpectralAlbedo provides a stub implementation that converts RGB albedo to spectral
// by using the average of the RGB components as a neutral spectral response
func (ns *nonSpectral) SpectralAlbedo(u float64, v float64, lambda float64, p *vec3.Vec3Impl) float64 {
	// This is a fallback that converts RGB to neutral spectral response
	// In practice, spectral materials should override this method
	return 0.5 // Default neutral response
}
