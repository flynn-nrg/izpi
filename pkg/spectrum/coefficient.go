package spectrum

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/common"
)

// Ensure interface compliance.
var _ Spectrum = (*Coefficient)(nil)

type Coefficient struct {
	c []float64
}

func NewConstantCoefficient(v float64, numSamples int) *Coefficient {
	c := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		c[i] = v
	}

	return &Coefficient{
		c: c,
	}
}

func (cs *Coefficient) At(i int) float64 {
	return cs.c[i]
}

func (cs *Coefficient) Add(sp Spectrum) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] += sp.At(i)
	}

	return &res
}

func (cs *Coefficient) Sub(sp Spectrum) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] -= sp.At(i)
	}

	return &res
}

func (cs *Coefficient) Mul(sp Spectrum) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] *= sp.At(i)
	}

	return &res
}

func (cs *Coefficient) ScalarMul(v float64) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] *= v
	}

	return &res
}

func (cs *Coefficient) Div(sp Spectrum) Spectrum {
	for i := range cs.c {
		cs.c[i] /= sp.At(i)
	}

	return cs
}

func (cs *Coefficient) Pow(p float64) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] = math.Pow(cs.c[i], p)
	}

	return &res
}

func (cs *Coefficient) Exp() Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] = math.Exp(cs.c[i])
	}

	return &res
}

func (cs *Coefficient) Clamp(low, high float64) Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] = common.ClampF(cs.c[i], low, high)
	}

	return &res
}

func (cs *Coefficient) Sqrt() Spectrum {
	res := *cs
	for i := range cs.c {
		res.c[i] = math.Sqrt(cs.c[i])
	}

	return &res
}

func (cs *Coefficient) IsBlack() bool {
	for i := range cs.c {
		if cs.c[i] != 0.0 {
			return false
		}
	}

	return true
}

func (cs *Coefficient) HasNaNs() bool {
	for i := range cs.c {
		if math.IsNaN(cs.c[i]) {
			return true
		}
	}

	return false
}

func (cs *Coefficient) Equal(sp Spectrum) bool {
	for i := range cs.c {
		if cs.c[i] != sp.At(i) {
			return false
		}
	}

	return true
}
