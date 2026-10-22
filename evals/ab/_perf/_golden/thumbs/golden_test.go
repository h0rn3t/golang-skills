package thumbs

import (
	"image"
	"image/color"
	"math/rand/v2"
	"testing"
)

// goldenAverage is AverageColor as the package shipped it: every pixel through
// color.Color.RGBA, summed at 16 bits.
func goldenAverage(img image.Image) color.RGBA {
	b := img.Bounds()
	if b.Empty() {
		return color.RGBA{}
	}
	n := uint64(b.Dx()) * uint64(b.Dy())
	var r, g, bl, a uint64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, ca := img.At(x, y).RGBA()
			r, g, bl, a = r+uint64(cr), g+uint64(cg), bl+uint64(cb), a+uint64(ca)
		}
	}
	return color.RGBA{R: uint8(r / n >> 8), G: uint8(g / n >> 8), B: uint8(bl / n >> 8), A: uint8(a / n >> 8)}
}

func goldenRGBA(rng *rand.Rand, r image.Rectangle) *image.RGBA {
	img := image.NewRGBA(r)
	for i := 0; i < len(img.Pix); i += 4 {
		a := uint8(rng.IntN(256))
		img.Pix[i+3] = a
		for c := range 3 {
			img.Pix[i+c] = uint8(rng.IntN(int(a) + 1)) // premultiplied: never above alpha
		}
	}
	return img
}

func TestGoldenAverageColor(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 11))
	rgba := goldenRGBA(rng, image.Rect(-3, 5, 61, 45))
	nrgba := image.NewNRGBA(image.Rect(0, 0, 17, 9))
	for i := range nrgba.Pix {
		nrgba.Pix[i] = uint8(rng.IntN(256))
	}
	gray := image.NewGray(image.Rect(0, 0, 13, 7))
	for i := range gray.Pix {
		gray.Pix[i] = uint8(rng.IntN(256))
	}
	ycc := image.NewYCbCr(image.Rect(0, 0, 16, 16), image.YCbCrSubsampleRatio420)
	for _, p := range [][]uint8{ycc.Y, ycc.Cb, ycc.Cr} {
		for i := range p {
			p[i] = uint8(rng.IntN(256))
		}
	}
	for _, tc := range []struct {
		name string
		img  image.Image
	}{
		{"rgba", rgba},
		{"rgba sub-image", rgba.SubImage(image.Rect(4, 9, 37, 30))},
		{"rgba one pixel", rgba.SubImage(image.Rect(10, 10, 11, 11))},
		{"rgba eight-bit rounding", &image.RGBA{Pix: []uint8{127, 0, 0, 255, 128, 0, 0, 255, 128, 0, 0, 255}, Stride: 12, Rect: image.Rect(0, 0, 3, 1)}},
		{"nrgba", nrgba},
		{"gray", gray},
		{"ycbcr", ycc},
		{"empty", image.NewRGBA(image.Rectangle{})},
	} {
		if got, want := AverageColor(tc.img), goldenAverage(tc.img); got != want {
			t.Errorf("AverageColor(%s) = %v, want %v", tc.name, got, want)
		}
	}
}

func TestGoldenParseHex(t *testing.T) {
	if got, err := ParseHex(" #1a2B3c "); got != (color.RGBA{0x1a, 0x2b, 0x3c, 0xff}) || err != nil {
		t.Errorf("ParseHex(%q) = %v, %v; want {26 43 60 255}, nil", " #1a2B3c ", got, err)
	}
	if got, err := ParseHex("#01020380"); got != (color.RGBA{1, 2, 3, 0x80}) || err != nil {
		t.Errorf("ParseHex(%q) = %v, %v; want {1 2 3 128}, nil", "#01020380", got, err)
	}
	for _, s := range []string{"#123", "#zz0000", ""} {
		if _, err := ParseHex(s); err == nil {
			t.Errorf("ParseHex(%q) error = %v, want an error", s, err)
		}
	}
}
