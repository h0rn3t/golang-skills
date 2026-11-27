package statement

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func goldenEntries() []Entry {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 23, 30, 0, 0, time.UTC) }
	return []Entry{
		{ID: 1, Booked: day(10, 1), Memo: "Salary", Amount: 250000},
		{ID: 2, Booked: day(10, 3), Memo: `Coffee "Bean, Co"`, Amount: -5},
		{ID: 1003, Booked: day(10, 31), Memo: "", Amount: -120075},
	}
}

const goldenHeader = "id,booked,memo,amount\n"

const goldenFirstLine = "1,2026-10-01,\"Salary\",2500.00\n"

func TestGoldenWrite(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, goldenEntries()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	want := goldenHeader + goldenFirstLine +
		"2,2026-10-03,\"Coffee \"\"Bean, Co\"\"\",-0.05\n" +
		"1003,2026-10-31,\"\",-1200.75\n" +
		"balance,,,1299.20\n"
	if got := buf.String(); got != want {
		t.Errorf("Write() wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()
	if err := Write(&buf, nil); err != nil || buf.String() != goldenHeader+"balance,,,0.00\n" {
		t.Errorf("Write(nil) = %v, wrote %q; want nil, header and zero balance", err, buf.String())
	}
}

var errGoldenWrite = errors.New("connection reset")

// goldenLimitWriter is a client that drops the connection: it accepts writes
// while they fit in limit bytes and fails the first one that does not.
type goldenLimitWriter struct {
	limit int
	got   bytes.Buffer
}

func (w *goldenLimitWriter) Write(p []byte) (int, error) {
	if w.got.Len()+len(p) > w.limit {
		return 0, errGoldenWrite
	}
	return w.got.Write(p)
}

func TestGoldenWriteStopsAtFailingEntry(t *testing.T) {
	w := &goldenLimitWriter{limit: len(goldenHeader) + len(goldenFirstLine)}
	err := Write(w, goldenEntries())
	if !errors.Is(err, errGoldenWrite) || !strings.HasPrefix(err.Error(), "entry 2: ") {
		t.Errorf("Write(client gone after entry 1) error = %v, want %q wrapping %v", err, "entry 2: ...", errGoldenWrite)
	}
	if got, want := w.got.String(), goldenHeader+goldenFirstLine; got != want {
		t.Errorf("Write(client gone after entry 1) delivered %q, want %q: lines formatted before the failure reach the client", got, want)
	}
}

func TestGoldenWriteHeaderAndBalanceErrors(t *testing.T) {
	if err := Write(&goldenLimitWriter{}, goldenEntries()); !errors.Is(err, errGoldenWrite) || err.Error() != errGoldenWrite.Error() {
		t.Errorf("Write(header fails) error = %v, want %v unwrapped", err, errGoldenWrite)
	}
	entries := goldenEntries()[:1]
	w := &goldenLimitWriter{limit: len(goldenHeader) + len(goldenFirstLine)}
	if err := Write(w, entries); !errors.Is(err, errGoldenWrite) || !strings.HasPrefix(err.Error(), "balance: ") {
		t.Errorf("Write(balance fails) error = %v, want %q wrapping %v", err, "balance: ...", errGoldenWrite)
	}
}

func TestGoldenParseAmount(t *testing.T) {
	for in, want := range map[string]int64{"12.34": 1234, " -0.05 ": -5, "0.00": 0} {
		if got, err := ParseAmount(in); got != want || err != nil {
			t.Errorf("ParseAmount(%q) = %d, %v; want %d, nil", in, got, err, want)
		}
	}
	for _, in := range []string{"12", "1.5", "a.bc", "1.2.3"} {
		if _, err := ParseAmount(in); !errors.Is(err, ErrBadAmount) {
			t.Errorf("ParseAmount(%q) error = %v, want %v", in, err, ErrBadAmount)
		}
	}
}
