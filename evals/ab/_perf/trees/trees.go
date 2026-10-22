// Package trees is the binary-trees benchmark: it builds complete binary
// trees bottom-up, walks them for a checksum, and reports one line per depth.
package trees

import "strconv"

type node struct {
	left, right *node
	item        uint32
}

// bottomUp builds a complete tree of the given depth whose items are derived
// from seed.
func bottomUp(seed uint32, depth int) *node {
	n := &node{item: seed}
	if depth > 0 {
		n.left = bottomUp(seed*2654435761+1, depth-1)
		n.right = bottomUp(seed*2246822519+3, depth-1)
	}
	return n
}

// check folds every item of the tree into one checksum.
func (n *node) check() uint32 {
	if n.left == nil {
		return n.item
	}
	return n.item ^ (n.left.check()*31 + n.right.check())
}

// Run reports the binary-trees workload for maxDepth (at least 6): a stretch
// tree one level deeper, then for every even depth from 4 the summed checksum
// of 2^(maxDepth-depth+4) trees seeded 1, 2, ..., and last a tree of maxDepth
// that was built before all of them and lived through them.
func Run(maxDepth int) []string {
	const minDepth = 4
	maxDepth = max(maxDepth, minDepth+2)

	stretch := maxDepth + 1
	lines := []string{formatLine("stretch tree", stretch, bottomUp(1, stretch).check())}

	longLived := bottomUp(1, maxDepth)
	for depth := minDepth; depth <= maxDepth; depth += 2 {
		iterations := 1 << (maxDepth - depth + minDepth)
		var sum uint32
		for i := 1; i <= iterations; i++ {
			sum += bottomUp(uint32(i), depth).check()
		}
		lines = append(lines, formatLine(strconv.Itoa(iterations)+" trees", depth, sum))
	}
	return append(lines, formatLine("long lived tree", maxDepth, longLived.check()))
}

// formatLine renders one report line.
func formatLine(label string, depth int, check uint32) string {
	s := ""
	s += label
	s += " of depth "
	s += strconv.Itoa(depth)
	s += "\t check: "
	s += strconv.FormatUint(uint64(check), 10)
	return s
}
