package trees

import (
	"slices"
	"testing"
)

func TestGoldenRun(t *testing.T) {
	cases := []struct {
		depth int
		want  []string
	}{
		{0, []string{"stretch tree of depth 7\t check: 2456820873", "64 trees of depth 4\t check: 506176288", "16 trees of depth 6\t check: 312197512", "long lived tree of depth 6\t check: 3977170857"}},
		{6, []string{"stretch tree of depth 7\t check: 2456820873", "64 trees of depth 4\t check: 506176288", "16 trees of depth 6\t check: 312197512", "long lived tree of depth 6\t check: 3977170857"}},
		{10, []string{"stretch tree of depth 11\t check: 770002985", "1024 trees of depth 4\t check: 3124795904", "256 trees of depth 6\t check: 1723703424", "64 trees of depth 8\t check: 1475841056", "16 trees of depth 10\t check: 458665224", "long lived tree of depth 10\t check: 2723927273"}},
	}
	for _, tc := range cases {
		if got := Run(tc.depth); !slices.Equal(got, tc.want) {
			t.Errorf("Run(%d) = %q, want %q", tc.depth, got, tc.want)
		}
	}
}
