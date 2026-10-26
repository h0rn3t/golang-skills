package zoning

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	goldenBase       = time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	goldenErrHistory = errors.New("history: connection reset")
	goldenErrSetZone = errors.New("zones: deadlock detected")
)

func goldenDay(n int) time.Time { return goldenBase.AddDate(0, 0, n) }

func goldenSpan(from, to int) Period { return Period{From: goldenDay(from), To: goldenDay(to)} }

func goldenSet(from, to int, zone string) string {
	return fmt.Sprintf("%s..%s=%s", goldenDay(from).Format("01-02"), goldenDay(to).Format("01-02"), zone)
}

// goldenStore serves fixed snapshots and records every call in order.
type goldenStore struct {
	snapshots   []Snapshot
	historyErr  error
	failSetAt   int // 1-based SetZone call that fails; 0 never
	historyArgs []Period
	sets        []string
}

func (s *goldenStore) History(_ context.Context, _ string, period Period) ([]Snapshot, error) {
	s.historyArgs = append(s.historyArgs, period)
	if s.historyErr != nil {
		return nil, s.historyErr
	}
	return s.snapshots, nil
}

func (s *goldenStore) SetZone(_ context.Context, _ string, period Period, zone *string) error {
	z := "<nil>"
	if zone != nil {
		z = *zone
	}
	s.sets = append(s.sets, fmt.Sprintf("%s..%s=%s", period.From.Format("01-02"), period.To.Format("01-02"), z))
	if s.failSetAt == len(s.sets) {
		return goldenErrSetZone
	}
	return nil
}

// goldenHistory is a match over days 2-5, a gap, a non-match over days 6-8,
// and a match that runs past the period's end.
var goldenHistory = []Snapshot{
	{Period: goldenSpan(2, 5), Props: map[string]any{"region": "north", "voltage": 220.0, "kind": "site"}},
	{Period: goldenSpan(6, 8), Props: map[string]any{"region": "south", "voltage": 220.0, "kind": "site"}},
	{Period: goldenSpan(8, 20), Props: map[string]any{"region": "north", "voltage": 220.0, "kind": "site"}},
}

var goldenMatch = map[string]string{"region": "north", "voltage": "220"}

func TestApply(t *testing.T) {
	tests := []struct {
		name        string
		snapshots   []Snapshot
		launch      time.Time
		req         Request
		wantOK      bool
		wantHistory []Period
		wantSets    []string
	}{
		{
			name:   "period ends before launch",
			launch: goldenDay(10),
			req:    Request{Point: "p1", Period: goldenSpan(0, 10), Zone: "Z1", Bind: true, Static: true},
		},
		{
			name:     "static bind starts at launch",
			launch:   goldenDay(3),
			req:      Request{Point: "p1", Period: goldenSpan(0, 10), Zone: "Z1", Bind: true, Static: true},
			wantOK:   true,
			wantSets: []string{goldenSet(3, 10, "Z1")},
		},
		{
			name:     "static unbind clears",
			launch:   goldenDay(0),
			req:      Request{Point: "p1", Period: goldenSpan(1, 4), Zone: "Z1", Static: true},
			wantOK:   true,
			wantSets: []string{goldenSet(1, 4, "<nil>")},
		},
		{
			name:        "bind sets matches and clears the rest",
			snapshots:   goldenHistory,
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 10), Zone: "Z1", Bind: true, Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 10)},
			wantSets: []string{
				goldenSet(0, 2, "<nil>"), goldenSet(2, 5, "Z1"), goldenSet(5, 6, "<nil>"),
				goldenSet(6, 8, "<nil>"), goldenSet(8, 10, "Z1"),
			},
		},
		{
			name:        "unbind leaves uncovered pieces alone",
			snapshots:   goldenHistory,
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 10), Zone: "Z1", Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 10)},
			wantSets:    []string{goldenSet(2, 5, "<nil>"), goldenSet(6, 8, "<nil>"), goldenSet(8, 10, "<nil>")},
		},
		{
			name:        "launch clips the first snapshot",
			snapshots:   goldenHistory,
			launch:      goldenDay(3),
			req:         Request{Point: "p1", Period: goldenSpan(0, 7), Zone: "Z1", Bind: true, Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(3, 7)},
			wantSets:    []string{goldenSet(3, 5, "Z1"), goldenSet(5, 6, "<nil>"), goldenSet(6, 7, "<nil>")},
		},
		{
			name:        "no history on bind clears the whole period",
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 4), Zone: "Z1", Bind: true, Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 4)},
			wantSets:    []string{goldenSet(0, 4, "<nil>")},
		},
		{
			name:        "no history on unbind touches nothing",
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 4), Zone: "Z1", Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 4)},
		},
		{
			name: "values compare as text",
			snapshots: []Snapshot{
				{Period: goldenSpan(0, 1), Props: map[string]any{"ratio": 0.4, "metered": true, "kind": "kiosk"}},
				{Period: goldenSpan(1, 2), Props: map[string]any{"ratio": 0.5, "metered": true, "kind": "kiosk"}},
				{Period: goldenSpan(2, 3), Props: map[string]any{"ratio": 0.4, "metered": false, "kind": "kiosk"}},
			},
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 3), Zone: "Z2", Bind: true, Match: map[string]string{"ratio": "0.4", "metered": "true", "kind": "kiosk"}},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 3)},
			wantSets:    []string{goldenSet(0, 1, "Z2"), goldenSet(1, 2, "<nil>"), goldenSet(2, 3, "<nil>")},
		},
		{
			name: "unrecorded kind reads as site",
			snapshots: []Snapshot{
				{Period: goldenSpan(0, 1), Props: map[string]any{"region": "north"}},
				{Period: goldenSpan(1, 2), Props: map[string]any{"region": "north", "kind": nil}},
				{Period: goldenSpan(2, 3), Props: map[string]any{"region": "north", "kind": "kiosk"}},
			},
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 3), Zone: "Z3", Bind: true, Match: map[string]string{"region": "north", "kind": "site"}},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 3)},
			wantSets:    []string{goldenSet(0, 1, "Z3"), goldenSet(1, 2, "Z3"), goldenSet(2, 3, "<nil>")},
		},
		{
			name: "unrecorded property does not match",
			snapshots: []Snapshot{
				{Period: goldenSpan(0, 1), Props: map[string]any{"region": "north"}},
				{Period: goldenSpan(1, 2), Props: map[string]any{"region": "north", "voltage": nil}},
			},
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(0, 2), Zone: "Z1", Bind: true, Match: goldenMatch},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(0, 2)},
			wantSets:    []string{goldenSet(0, 1, "<nil>"), goldenSet(1, 2, "<nil>")},
		},
		{
			name:        "empty match matches every snapshot",
			snapshots:   goldenHistory[:1],
			launch:      goldenDay(0),
			req:         Request{Point: "p1", Period: goldenSpan(2, 5), Zone: "Z1", Bind: true},
			wantOK:      true,
			wantHistory: []Period{goldenSpan(2, 5)},
			wantSets:    []string{goldenSet(2, 5, "Z1")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &goldenStore{snapshots: tt.snapshots}
			ok, err := Apply(t.Context(), st, tt.launch, tt.req)
			if err != nil {
				t.Fatalf("Apply() error = %v", err)
			}
			if ok != tt.wantOK {
				t.Errorf("Apply() = %v, want %v", ok, tt.wantOK)
			}
			if !reflect.DeepEqual(st.historyArgs, tt.wantHistory) {
				t.Errorf("History calls = %v, want %v", st.historyArgs, tt.wantHistory)
			}
			if !reflect.DeepEqual(st.sets, tt.wantSets) {
				t.Errorf("SetZone calls =\n  %q\nwant\n  %q", st.sets, tt.wantSets)
			}
		})
	}
}

func TestApplyBadValueFailsBeforeAnyWrite(t *testing.T) {
	st := &goldenStore{snapshots: []Snapshot{
		{Period: goldenSpan(0, 1), Props: map[string]any{"region": "north", "voltage": 220.0}},
		{Period: goldenSpan(1, 2), Props: map[string]any{"region": []any{"north"}, "voltage": 220.0}},
	}}
	req := Request{Point: "meter-77", Period: goldenSpan(0, 2), Zone: "Z1", Bind: true, Match: goldenMatch}
	ok, err := Apply(t.Context(), st, goldenDay(0), req)
	if err == nil {
		t.Fatalf("Apply() error = nil, want an error")
	}
	if ok {
		t.Errorf("Apply() = true on error")
	}
	if msg := err.Error(); !strings.Contains(msg, "region") || !strings.Contains(msg, "meter-77") {
		t.Errorf("Apply() error = %q, want it to name region and meter-77", msg)
	}
	if len(st.sets) != 0 {
		t.Errorf("SetZone calls = %q, want none before the error", st.sets)
	}
}

func TestApplyFailures(t *testing.T) {
	t.Run("history", func(t *testing.T) {
		st := &goldenStore{historyErr: goldenErrHistory}
		req := Request{Point: "p1", Period: goldenSpan(0, 4), Zone: "Z1", Bind: true, Match: goldenMatch}
		ok, err := Apply(t.Context(), st, goldenDay(0), req)
		if !errors.Is(err, goldenErrHistory) || ok {
			t.Errorf("Apply() = %v, %v, want false and an error wrapping %v", ok, err, goldenErrHistory)
		}
		if len(st.sets) != 0 {
			t.Errorf("SetZone calls = %q, want none", st.sets)
		}
	})
	t.Run("set zone stops at the first failure", func(t *testing.T) {
		st := &goldenStore{snapshots: goldenHistory, failSetAt: 2}
		req := Request{Point: "p1", Period: goldenSpan(0, 10), Zone: "Z1", Bind: true, Match: goldenMatch}
		ok, err := Apply(t.Context(), st, goldenDay(0), req)
		if !errors.Is(err, goldenErrSetZone) || ok {
			t.Errorf("Apply() = %v, %v, want false and an error wrapping %v", ok, err, goldenErrSetZone)
		}
		if len(st.sets) != 2 {
			t.Errorf("SetZone calls = %q, want 2", st.sets)
		}
	})
}
