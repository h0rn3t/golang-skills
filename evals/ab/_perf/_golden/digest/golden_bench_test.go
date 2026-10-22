package digest

import "testing"

func BenchmarkGoldenName(b *testing.B) {
	blob := make([]byte, 4<<20)
	for i := range blob {
		blob[i] = byte(i * 31)
	}
	for b.Loop() {
		if len(Name(blob)) != 71 {
			b.Fatal("Name returned a malformed name")
		}
	}
}
