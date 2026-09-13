package roster

import (
	"fmt"
	"sort"
	"strings"
)

// Longest returns the member who joined earliest and false when members is
// empty. On a tie the first such member wins.
func Longest(members []Member) (Member, bool) {
	if len(members) == 0 {
		return Member{}, false
	}
	best := members[0]
	for i := 0; i < len(members); i++ {
		if members[i].Joined < best.Joined {
			best = members[i]
		}
	}
	return best, true
}

// Teams returns every team that appears in members, sorted, each once.
func Teams(members []Member) []string {
	seen := map[string]bool{}
	var teams []string
	for _, m := range members {
		m := m
		if !seen[m.Team] {
			seen[m.Team] = true
			teams = append(teams, m.Team)
		}
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i] < teams[j] })
	return teams
}

// SplitHandle splits a "team:name" handle into its two halves. A handle
// without a colon is all team and no name.
func SplitHandle(handle string) (team, name string) {
	i := strings.Index(handle, ":")
	if i < 0 {
		return handle, ""
	}
	return handle[:i], handle[i+1:]
}

// Describe renders any value the way log lines in this package always have.
func Describe(v interface{}) string {
	return fmt.Sprintf("roster: %v", v)
}
