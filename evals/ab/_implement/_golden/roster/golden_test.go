package roster

import (
	"slices"
	"testing"
)

var goldenMembers = []Member{
	{Name: "Mara", Team: "ops", Active: true, Joined: 2019},
	{Name: "Ivo", Team: "ops", Active: false, Joined: 2017},
	{Name: "Zoe", Team: "dev", Active: true, Joined: 2021},
	{Name: "Mara", Team: "ops", Active: true, Joined: 2023},
	{Name: "Ana", Team: "ops", Active: true, Joined: 2020},
	{Name: "Kai", Team: "qa", Active: false, Joined: 2018},
}

func TestActiveNamesSortedUniqueActiveOnly(t *testing.T) {
	got := ActiveNames(goldenMembers, "ops")
	if want := []string{"Ana", "Mara"}; !slices.Equal(got, want) {
		t.Fatalf("ActiveNames(ops) = %q, want %q", got, want)
	}
	if got := ActiveNames(goldenMembers, "qa"); len(got) != 0 {
		t.Errorf("ActiveNames(qa) = %q, want no names: its only member is inactive", got)
	}
	if got := ActiveNames(nil, "ops"); len(got) != 0 {
		t.Errorf("ActiveNames(nil) = %q, want no names", got)
	}
}

func TestHeadCountActivePerTeam(t *testing.T) {
	got := HeadCount(goldenMembers)
	want := map[string]int{"ops": 3, "dev": 1}
	if len(got) != len(want) {
		t.Fatalf("HeadCount = %v, want %v", got, want)
	}
	for team, n := range want {
		if got[team] != n {
			t.Errorf("HeadCount[%q] = %d, want %d", team, got[team], n)
		}
	}
	if _, ok := got["qa"]; ok {
		t.Errorf("HeadCount has %q, want it absent: no active member", "qa")
	}
}

func TestOnTeamCountsInactiveMembers(t *testing.T) {
	if !OnTeam(goldenMembers, "qa") {
		t.Error("OnTeam(qa) = false, want true: an inactive member is still on the team")
	}
	if OnTeam(goldenMembers, "design") {
		t.Error("OnTeam(design) = true, want false")
	}
}

// The legacy functions are in production use; their behavior is pinned so a
// change to this package cannot move them.
func TestLegacyBehaviorUnchanged(t *testing.T) {
	if got, ok := Longest(goldenMembers); !ok || got.Name != "Ivo" {
		t.Errorf("Longest = %+v, %v; want Ivo, true", got, ok)
	}
	if _, ok := Longest(nil); ok {
		t.Error("Longest(nil) ok = true, want false")
	}
	if got, want := Teams(goldenMembers), []string{"dev", "ops", "qa"}; !slices.Equal(got, want) {
		t.Errorf("Teams = %q, want %q", got, want)
	}
	if team, name := SplitHandle("ops:Mara"); team != "ops" || name != "Mara" {
		t.Errorf("SplitHandle(ops:Mara) = %q, %q", team, name)
	}
	if team, name := SplitHandle("ops"); team != "ops" || name != "" {
		t.Errorf("SplitHandle(ops) = %q, %q; want ops and no name", team, name)
	}
	if got := Describe(42); got != "roster: 42" {
		t.Errorf("Describe(42) = %q, want %q", got, "roster: 42")
	}
}
