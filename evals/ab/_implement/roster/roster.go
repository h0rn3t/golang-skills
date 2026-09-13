// Package roster answers questions about the members of a team.
//
// legacy.go has been in production since the package was written and is not
// part of this change; the functions below are new.
package roster

// Member is one person on the roster.
type Member struct {
	Name   string
	Team   string
	Active bool
	Joined int // the year the member joined
}

// ActiveNames returns the names of the active members of team, sorted and
// without duplicates. A team with no active members yields no names.
func ActiveNames(members []Member, team string) []string {
	panic("not implemented")
}

// HeadCount returns how many active members each team has, keyed by team.
// A team with no active member is absent from the result.
func HeadCount(members []Member) map[string]int {
	panic("not implemented")
}

// OnTeam reports whether any member, active or not, is on team.
func OnTeam(members []Member, team string) bool {
	panic("not implemented")
}
