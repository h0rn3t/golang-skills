package probe

import (
	"context"
	"os"
	"testing"
)

func TestCleanupContext(t *testing.T) {
	t.Cleanup(func() {
		if err := context.WithoutCancel(t.Context()).Err(); err != nil {
			t.Error(err)
		}
	})
}

func TestMain(m *testing.M) {
	if err := context.Background().Err(); err != nil {
		os.Exit(1)
	}
	m.Run()
}
