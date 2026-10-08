package quota

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("QUOTA_TEST_DB") == "" {
		log.Print("QUOTA_TEST_DB is not set; the integration database is unavailable")
		return
	}
	m.Run()
}

func TestWorkspaceLimit(t *testing.T) {
	for _, tt := range []struct {
		plan Plan
		want int
	}{
		{Free, 1},
		{Team, 10},
		{Enterprise, 100},
	} {
		if got := workspaceLimit(tt.plan); got != tt.want {
			t.Errorf("workspaceLimit(%v) = %d, want %d", tt.plan, got, tt.want)
		}
	}
}

func TestInvoice(t *testing.T) {
	if got := Invoice(3, "EUR"); got != 3300 {
		t.Errorf("Invoice(3, EUR) = %d, want 3300", got)
	}
}
