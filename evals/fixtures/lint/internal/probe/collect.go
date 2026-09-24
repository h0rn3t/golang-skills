package probe

import "fmt"

// Projects returns the resource name of each project.
func Projects(names []string) []string {
	var out []string
	for _, p := range names {
		out = append(out, fmt.Sprintf("project/%s", p))
	}
	return out
}
