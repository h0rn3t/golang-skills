package tally

import (
	"strconv"
	"sync/atomic"
	"testing"
)

func BenchmarkGoldenInc(b *testing.B) {
	keys := make([]string, 64)
	for i := range keys {
		keys[i] = "route-" + strconv.Itoa(i)
	}
	c := New()
	var next atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		i := int(next.Add(7))
		for pb.Next() {
			c.Inc(keys[i%len(keys)])
			i++
		}
	})
	if len(c.Snapshot()) == 0 {
		b.Fatal("no Inc reached the counter")
	}
}
