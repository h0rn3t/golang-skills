package probe

// Longest names a local after a predeclared identifier: go-naming says not to.
func Longest(words []string) int {
	max := 0
	for _, w := range words {
		if n := len(w); n > max {
			max = n
		}
	}
	return max
}
