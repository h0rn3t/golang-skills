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

func TestResultTextCLIError(t *testing.T) {
	tests := []struct {
		name, out string
		wantErr   bool
	}{
		{name: "usage limit", out: `{"type":"result","subtype":"success","is_error":true,"result":"You've hit your session limit"}`, wantErr: true},
		{name: "usage limit in array", out: `[{"type":"system"},{"type":"result","is_error":true,"result":"You've hit your session limit"}]`, wantErr: true},
		{name: "out of turns", out: `{"type":"result","subtype":"error_max_turns","is_error":true,"result":"partial answer"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := resultText([]byte(tt.out)); (err != nil) != tt.wantErr {
				t.Errorf("resultText(%s) error = %v, want error %t", tt.out, err, tt.wantErr)
			}
		})
	}
}

func TestStreamError(t *testing.T) {
	limit := "{\"type\":\"system\"}\n{\"type\":\"result\",\"is_error\":true,\"result\":\"You've hit your session limit\"}\n"
	if streamError([]byte(limit)) == "" {
		t.Errorf("streamError(%q) = \"\", want the CLI's error", limit)
	}
	ok := "{\"type\":\"result\",\"is_error\":false,\"result\":\"done\"}\n"
	if got := streamError([]byte(ok)); got != "" {
		t.Errorf("streamError(%q) = %q, want \"\"", ok, got)
	}
}

func TestTriggered(t *testing.T) {
	tests := []struct {
		name          string
		want, notWant []string
		got           []string
		pass          bool
	}{
		{name: "expected fired", want: []string{"go-context"}, got: []string{"go-context", "go-style-core"}, pass: true},
		{name: "expected missing", want: []string{"go-context"}, got: []string{"go-concurrency"}},
		{name: "neighbor fired too", want: []string{"go-context"}, notWant: []string{"go-concurrency"}, got: []string{"go-concurrency", "go-context"}},
		{name: "neighbor quiet", want: []string{"go-context"}, notWant: []string{"go-concurrency"}, got: []string{"go-context", "go-style-core"}, pass: true},
		{name: "negative quiet", pass: true},
		{name: "negative fired", got: []string{"go-code"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := map[string]bool{}
			for _, name := range tt.got {
				got[name] = true
			}
			if pass := triggered(tt.want, tt.notWant, got); pass != tt.pass {
				t.Errorf("triggered(%q, %q, %q) = %t, want %t", tt.want, tt.notWant, tt.got, pass, tt.pass)
			}
		})
	}
}

func TestTriggerPrecision(t *testing.T) {
	results := []triggerResult{
		{Want: []string{"go-context"}, Got: []string{"go-concurrency", "go-context"}},
		{Want: []string{"go-http", "go-security"}, Got: []string{"go-http", "go-security"}},
		{Want: nil, Got: []string{"go-code"}},
		{Want: []string{"go-testing"}},
		// An errored session's partial transcript is left out of the pool.
		{Want: []string{"go-naming"}, Got: []string{"go-code", "go-packages"}, Err: "timed out"},
	}
	// 1 of 2, 2 of 2, 0 of 1, nothing fired: 3 of 5.
	if got := triggerPrecision(results); got == nil || *got != 0.6 {
		t.Errorf("triggerPrecision(results) = %v, want 0.6", got)
	}
	if got := triggerPrecision([]triggerResult{{Want: []string{"go-testing"}}}); got != nil {
		t.Errorf("triggerPrecision(nothing fired) = %v, want nil", *got)
	}
}

func TestQualityToolArgs(t *testing.T) {
	tests := []struct {
		name string
		o    options
		want []string
	}{
		{name: "read only", want: []string{"--tools", "Read", "--allowed-tools", "Read"}},
		{name: "go doc", o: options{goDoc: true}, want: []string{"--tools", "Read,Bash", "--allowed-tools", "Read,Bash(go doc:*)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualityToolArgs(tt.o, "Read"); !slices.Equal(got, tt.want) {
				t.Errorf("qualityToolArgs(goDoc=%t, \"Read\") = %q, want %q", tt.o.goDoc, got, tt.want)
			}
		})
	}
}
