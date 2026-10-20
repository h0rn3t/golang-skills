package spool

import (
	"bytes"
	"errors"
	"testing"
)

type goldenWriter struct {
	writes [][]byte
	err    error
}

func (g *goldenWriter) Write(p []byte) (int, error) {
	if g.err != nil {
		return 0, g.err
	}
	g.writes = append(g.writes, bytes.Clone(p))
	return len(p), nil
}

var goldenRows = []Row{
	{Name: "cpu", Values: []float64{0.5, 1.25}},
	{Name: "mem", Values: []float64{3.14159}, Digits: 4},
	{Name: "idle"},
}

const goldenText = "cpu 0.50 1.25 = 1.75\nmem 3.1416 = 3.14\nidle = 0.00\n"

// TestGoldenRender pins the text, the per-row precision override, the
// fallback to the report's precision for a row without Digits, and the total
// that keeps the report's precision even on a row that overrides it.
func TestGoldenRender(t *testing.T) {
	var out bytes.Buffer
	if err := Render(&out, goldenRows, 2); err != nil {
		t.Fatalf("Render(rows, 2) error = %v, want nil", err)
	}
	if got := out.String(); got != goldenText {
		t.Errorf("Render(rows, 2) wrote %q, want %q", got, goldenText)
	}
}

// TestGoldenRenderFlushesBuffer pins that a writer from NewBuffer reaches its
// destination in one write once Render returns.
func TestGoldenRenderFlushesBuffer(t *testing.T) {
	dst := &goldenWriter{}
	if err := Render(NewBuffer(dst), goldenRows, 2); err != nil {
		t.Fatalf("Render(NewBuffer(dst), rows, 2) error = %v, want nil", err)
	}
	if len(dst.writes) != 1 || string(dst.writes[0]) != goldenText {
		t.Errorf("Render(NewBuffer(dst), rows, 2) wrote %q, want one write of %q", dst.writes, goldenText)
	}
}

func TestGoldenRenderErrors(t *testing.T) {
	if err := Render(&bytes.Buffer{}, goldenRows, -1); err == nil || err.Error() != "spool: negative precision -1" {
		t.Errorf("Render(rows, -1) error = %v, want spool: negative precision -1", err)
	}
	nameless := []Row{{Name: "a"}, {}}
	if err := Render(&bytes.Buffer{}, nameless, 1); err == nil || err.Error() != "spool: row 1 has no name" {
		t.Errorf("Render(nameless, 1) error = %v, want spool: row 1 has no name", err)
	}

	boom := errors.New("disk full")
	err := Render(&goldenWriter{err: boom}, goldenRows, 2)
	if !errors.Is(err, boom) || err.Error() != `spool: write row "cpu": disk full` {
		t.Errorf("Render(failing writer) error = %v, want spool: write row \"cpu\": disk full", err)
	}
	err = Render(NewBuffer(&goldenWriter{err: boom}), goldenRows, 2)
	if !errors.Is(err, boom) || err.Error() != "spool: flush: disk full" {
		t.Errorf("Render(NewBuffer(failing writer)) error = %v, want spool: flush: disk full", err)
	}
}
