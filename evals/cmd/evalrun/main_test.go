package main

import (
	"slices"
	"testing"
)

func TestSelectQuality(t *testing.T) {
	evals := []qualityEval{
		{ID: 1, Set: "train"},
		{ID: 2, Set: "validation"},
		{ID: 3, Set: "validation"},
	}
	tests := []struct {
		name, set, ids string
		want           []int
		wantErr        bool
	}{
		{name: "set", set: "validation", want: []int{2, 3}},
		{name: "all", set: "all", want: []int{1, 2, 3}},
		{name: "ids ignore the set", set: "validation", ids: "3, 1", want: []int{3, 1}},
		{name: "unknown id", set: "all", ids: "2,9", wantErr: true},
		{name: "not a number", set: "all", ids: "2,x", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectQuality(evals, tt.set, tt.ids)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("selectQuality(%q, %q) error = %v, want error %t", tt.set, tt.ids, err, tt.wantErr)
			}
			var ids []int
			for _, ev := range got {
				ids = append(ids, ev.ID)
			}
			if !slices.Equal(ids, tt.want) {
				t.Errorf("selectQuality(%q, %q) = %v, want %v", tt.set, tt.ids, ids, tt.want)
			}
		})
	}
}

func TestResultTextCost(t *testing.T) {
	tests := []struct {
		name, out string
		text      string
		cost      float64
	}{
		{name: "object", out: `{"result":"ok","total_cost_usd":0.25}`, text: "ok", cost: 0.25},
		{name: "array", out: `[{"type":"system"},{"result":"done","total_cost_usd":1.5}]`, text: "done", cost: 1.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, cost, err := resultText([]byte(tt.out))
			if err != nil || text != tt.text || cost != tt.cost {
				t.Errorf("resultText(%s) = %q, %v, %v; want %q, %v, nil", tt.out, text, cost, err, tt.text, tt.cost)
			}
		})
	}
}
