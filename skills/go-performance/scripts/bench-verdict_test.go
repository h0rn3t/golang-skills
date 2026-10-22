package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExactU(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n1, n2, u int
		want      float64
	}{
		{1, 1, 0, 1},
		{3, 3, 0, 2.0 / 20},
		{3, 3, 9, 2.0 / 20},
		{3, 3, 4, 1},
		{10, 10, 0, 2.0 / 184756},
		{2, 3, 1, 4.0 / 10},
	}
	for _, tc := range cases {
		if got := exactU(tc.n1, tc.n2, tc.u); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("exactU(%d, %d, %d) = %g, want %g", tc.n1, tc.n2, tc.u, got, tc.want)
		}
	}
}

func TestUTestTies(t *testing.T) {
	t.Parallel()
	seven, three := repeat(7, 10), repeat(3, 10)
	if p := uTest(seven, three); p >= 0.001 {
		t.Errorf("uTest(10×7, 10×3) = %g, want < 0.001", p)
	}
	if p := uTest(seven, repeat(7, 10)); p != 1 {
		t.Errorf("uTest(10×7, 10×7) = %g, want 1", p)
	}
}

func TestCIPct(t *testing.T) {
	t.Parallel()
	ten := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := ciPct(ten, median(ten))
	if want := 3.5 / 5.5 * 100; got == nil || math.Abs(*got-want) > 1e-9 {
		t.Errorf("ciPct(1..10) = %v, want %g", got, want)
	}
	if got := ciPct(ten[:5], median(ten[:5])); got != nil {
		t.Errorf("ciPct(1..5) = %g, want nil: five samples have no 95%% interval", *got)
	}
	if got := ciPct(repeat(0, 10), 0); got == nil || *got != 0 {
		t.Errorf("ciPct(10×0) = %v, want 0", got)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	in := "goos: linux\nBenchmarkJoin-8   \t 1000\t      1500 ns/op\t     256 B/op\t       3 allocs/op\t 12.5 MB/s\nPASS\nok  \tx\t0.1s\n"
	s, err := parse(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := []key{{"BenchmarkJoin-8", unitSec}, {"BenchmarkJoin-8", unitBytes}, {"BenchmarkJoin-8", unitAllocs}, {"BenchmarkJoin-8", "MB/s"}}
	if fmt.Sprint(s.order) != fmt.Sprint(want) {
		t.Fatalf("parse order = %v, want %v", s.order, want)
	}
	if v := s.vals[want[0]][0]; math.Abs(v-1.5e-6) > 1e-15 {
		t.Errorf("parse ns/op = %g sec/op, want 1.5e-06", v)
	}
}

func TestVerdict(t *testing.T) {
	t.Parallel()
	tight := []float64{100, 101, 99, 100, 102, 98, 100, 101, 99, 100}
	cases := []struct {
		name               string
		baseNs, newNs      []float64
		baseAlloc, newAllc float64
		want               string
		exit               int
	}{
		{"faster", tight, scale(tight, 0.5), 3, 3, verdictKeep, 0},
		{"fewer allocations", tight, tight, 7, 3, verdictKeep, 0},
		{"more allocations", tight, scale(tight, 0.5), 3, 7, verdictRevert, 1},
		{"allocations up by less than the threshold", tight, scale(tight, 0.1), 6000, 6007, verdictKeep, 0},
		{"slower", tight, scale(tight, 1.5), 3, 3, verdictRevert, 1},
		{"unchanged", tight, tight, 3, 3, verdictRevert, 1},
		{"spread too wide", []float64{100, 60, 140, 80, 120, 70, 130, 90, 110, 100}, []float64{95, 65, 135, 85, 115, 75, 125, 95, 105, 100}, 3, 3, verdictNoisy, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			base := writeRuns(t, dir, "base.txt", tc.baseNs, tc.baseAlloc)
			changed := writeRuns(t, dir, "new.txt", tc.newNs, tc.newAllc)
			var stdout, stderr bytes.Buffer
			exit := run([]string{"--json", base, changed}, &stdout, &stderr)
			var rep report
			if err := json.Unmarshal(stdout.Bytes(), &rep); err != nil {
				t.Fatalf("run(--json) output is not a report: %v\n%s%s", err, stdout.String(), stderr.String())
			}
			if rep.Verdict != tc.want || exit != tc.exit {
				t.Errorf("run(%s) = verdict %q exit %d, want %q exit %d (reason %q)", tc.name, rep.Verdict, exit, tc.want, tc.exit, rep.Reason)
			}
		})
	}
}

func TestRunUsage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	few := writeRuns(t, dir, "few.txt", []float64{1, 2, 3}, 0)
	var out bytes.Buffer
	for _, args := range [][]string{
		{few},
		{"--threshold", "0", few, few},
		{few, few},
		{filepath.Join(dir, "missing.txt"), few},
	} {
		if exit := run(args, &out, &out); exit != 2 {
			t.Errorf("run(%q) exit = %d, want 2\n%s", args, exit, out.String())
		}
	}
}

func repeat(v float64, n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = v
	}
	return s
}

func scale(s []float64, f float64) []float64 {
	out := make([]float64, len(s))
	for i, v := range s {
		out[i] = v * f
	}
	return out
}

func writeRuns(t *testing.T, dir, name string, ns []float64, allocs float64) string {
	t.Helper()
	var b strings.Builder
	for _, v := range ns {
		fmt.Fprintf(&b, "BenchmarkJoin-8\t1000\t%g ns/op\t64 B/op\t%g allocs/op\n", v, allocs)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
