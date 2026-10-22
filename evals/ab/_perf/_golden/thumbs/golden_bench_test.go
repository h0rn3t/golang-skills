package thumbs

import (
	"image"
	"math/rand/v2"
	"testing"
)

// BenchmarkGoldenAverageColor runs one request per goroutine on every CPU, as
// the gallery service does: a speedup that spends more CPUs on one request has
// none to spend.
func BenchmarkGoldenAverageColor(b *testing.B) {
	img := goldenRGBA(rand.New(rand.NewPCG(3, 5)), image.Rect(0, 0, 256, 256))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c := AverageColor(img); c.A == 0 {
				b.Error("AverageColor of an opaque-ish image is transparent")
			}
		}
	})
}
