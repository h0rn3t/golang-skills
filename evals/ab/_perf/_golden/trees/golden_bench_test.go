package trees

import "testing"

var goldenSink []string

func BenchmarkGoldenRun(b *testing.B) {
	for b.Loop() {
		goldenSink = Run(12)
	}
}
