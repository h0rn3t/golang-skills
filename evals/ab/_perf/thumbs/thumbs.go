// Package thumbs computes the placeholder color the gallery shows while a
// thumbnail loads.
package thumbs

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

// AverageColor returns the mean color of img. Each channel is the sum of the
// 16-bit premultiplied values color.Color.RGBA reports for every pixel,
// divided by the pixel count and shifted down to 8 bits. An empty image is
// transparent black.
func AverageColor(img image.Image) color.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return color.RGBA{}
	}
	n := uint64(w) * uint64(h)
	var r, g, bl, a uint64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, ca := img.At(x, y).RGBA()
			r += uint64(cr)
			g += uint64(cg)
			bl += uint64(cb)
			a += uint64(ca)
		}
	}
	return color.RGBA{R: uint8(r / n >> 8 & 0xff), G: uint8(g / n >> 8 & 0xff), B: uint8(bl / n >> 8 & 0xff), A: uint8(a / n >> 8 & 0xff)}
}

// ParseHex reads the gallery's configured fallback color, "#rrggbb" or
// "#rrggbbaa", once at startup.
func ParseHex(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	c := color.RGBA{A: 0xff}
	var err error
	switch len(s) {
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 8:
		_, err = fmt.Sscanf(s, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	default:
		err = fmt.Errorf("color %q: want 6 or 8 hex digits", s)
	}
	if err != nil {
		return color.RGBA{}, fmt.Errorf("parse color %q: %w", s, err)
	}
	return c, nil
}
