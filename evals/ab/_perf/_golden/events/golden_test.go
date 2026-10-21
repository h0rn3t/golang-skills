package events

import (
	"bytes"
	"errors"
	"slices"
	"testing"
	"time"
)

func goldenEvents() []Event {
	kyiv := time.FixedZone("EET", 2*60*60)
	return []Event{
		{ID: 1, Kind: "login", Actor: "Ann", At: time.Date(2026, 10, 21, 9, 30, 0, 123456789, kyiv), Tags: []string{"web", "mfa"}, Amount: 0},
		{ID: 2, Kind: "payment", Actor: "<b>Bob & Co</b>", At: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Tags: nil, Amount: 1234567.89},
		{ID: -3, Kind: "refund", Actor: "ÉVA", At: time.Date(2025, 12, 31, 23, 59, 59, 1000, kyiv), Tags: []string{"z", "a", "m"}, Amount: 12.5},
	}
}

func TestGoldenWriteAll(t *testing.T) {
	events := goldenEvents()
	var buf bytes.Buffer
	if err := WriteAll(&buf, events); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}
	want := "{\"actor\":\"ann\",\"amount\":0,\"at\":\"2026-10-21T07:30:00.123456789Z\",\"id\":1,\"kind\":\"login\",\"tags\":[\"mfa\",\"web\"]}\n" +
		"{\"actor\":\"\\u003cb\\u003ebob \\u0026 co\\u003c/b\\u003e\",\"amount\":1234567.89,\"at\":\"2026-01-02T03:04:05Z\",\"id\":2,\"kind\":\"payment\",\"tags\":[]}\n" +
		"{\"actor\":\"éva\",\"amount\":12.5,\"at\":\"2025-12-31T21:59:59.000001Z\",\"id\":-3,\"kind\":\"refund\",\"tags\":[\"a\",\"m\",\"z\"]}\n"
	if got := buf.String(); got != want {
		t.Errorf("WriteAll() wrote\n%q\nwant\n%q", got, want)
	}
	if !slices.Equal(events[2].Tags, []string{"z", "a", "m"}) {
		t.Errorf("WriteAll() reordered the caller's tags: %q", events[2].Tags)
	}
	buf.Reset()
	if err := WriteAll(&buf, nil); err != nil || buf.Len() != 0 {
		t.Errorf("WriteAll(nil) = %v, wrote %q; want nil, nothing", err, buf.String())
	}
}

var errGoldenWrite = errors.New("disk full")

type goldenFailingWriter struct{}

func (goldenFailingWriter) Write([]byte) (int, error) { return 0, errGoldenWrite }

func TestGoldenWriteAllError(t *testing.T) {
	if err := WriteAll(goldenFailingWriter{}, goldenEvents()); !errors.Is(err, errGoldenWrite) {
		t.Errorf("WriteAll(failing writer) = %v, want %v", err, errGoldenWrite)
	}
}

func TestGoldenParseKind(t *testing.T) {
	for in, want := range map[string]string{" Login ": "login", "REFUND": "refund", "export": "export"} {
		if got, err := ParseKind(in); got != want || err != nil {
			t.Errorf("ParseKind(%q) = %q, %v; want %q, nil", in, got, err, want)
		}
	}
	if _, err := ParseKind("delete"); !errors.Is(err, ErrUnknownKind) {
		t.Errorf("ParseKind(%q) error = %v, want %v", "delete", err, ErrUnknownKind)
	}
}
