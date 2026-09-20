// Package floatcolor implements a float32-based colour model.
package floatcolor

import "image/color"

var _ (color.Color) = (*Float32NRGBA)(nil)

// Float32NRGBAModel implements the colour model for Float32RGBA colour data.
var Float32NRGBAModel color.Model = color.ModelFunc(float32NrgbaModel)

// Float32NRGBA represents a non-alpha-premultiplied float32 color.
type Float32NRGBA struct {
	R float32
	G float32
	B float32
	A float32
}

func (c Float32NRGBA) RGBA() (r, g, b, a uint32) {
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

func float32NrgbaModel(c color.Color) color.Color {
	if _, ok := c.(Float32NRGBA); ok {
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
