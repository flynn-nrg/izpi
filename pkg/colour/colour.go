// Package colour implements a float64-based colour model.
package colour

import "image/color"

var _ (color.Color) = (*FloatNRGBA)(nil)

// FloatNRGBAModel implements the colour model for FloatRGBA colour data.
var FloatNRGBAModel color.Model = color.ModelFunc(floatNrgbaModel)

// FloatNRGBA represents a non-alpha-premultiplied float64 color.
type FloatNRGBA struct {
	R float64
	G float64
	B float64
	A float64
}

func (c FloatNRGBA) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * 255.0)
	r |= r << 8
	r *= uint32(c.A * 255.0)
	r /= 0xff
	g = uint32(c.G * 255.0)
	g |= g << 8
	g *= uint32(c.A * 255.0)
	g /= 0xff
	b = uint32(c.B * 255.0)
	b |= b << 8
	b *= uint32(c.A * 255.0)
	b /= 0xff
	a = uint32(c.A * 255.0)
	a |= a << 8
	return
}

func floatNrgbaModel(c color.Color) color.Color {
	if _, ok := c.(FloatNRGBA); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xff}
	}
	if a == 0 {
		return color.NRGBA{0, 0, 0, 0}
	}
	// Since Color.RGBA returns an alpha-premultiplied color, we should have r <= a && g <= a && b <= a.
	r = (r * 0xffff) / a
	g = (g * 0xffff) / a
	b = (b * 0xffff) / a
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
