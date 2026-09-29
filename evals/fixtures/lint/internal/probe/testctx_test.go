package probe

import (
	"context"
	"testing"
)

func TestBackgroundContext(t *testing.T) {
	if err := context.Background().Err(); err != nil {
		t.Fatal(err)
	}
	if err := context.TODO().Err(); err != nil {
		t.Fatal(err)
	}
}
