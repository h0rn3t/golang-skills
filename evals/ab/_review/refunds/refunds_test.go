package refunds

import (
	"context"
	"testing"
)

type recordingGateway struct{ refunded []int64 }

func (g *recordingGateway) Refund(_ context.Context, _ string, amount int64) error {
	g.refunded = append(g.refunded, amount)
	return nil
}

func TestRefund(t *testing.T) {
	store := NewStore()
	if err := store.Put(t.Context(), Payment{ID: "p1", Amount: 1000, Currency: "EUR"}); err != nil {
		t.Fatal(err)
	}
	gw := &recordingGateway{}
	svc := NewService(store, gw)

	if err := svc.Refund(t.Context(), "p1", 600); err != nil {
		t.Fatalf("Refund(p1, 600) = %v, want nil", err)
	}
	if err := svc.Refund(t.Context(), "p1", 500); err == nil {
		t.Fatal("Refund(p1, 500) with 400 left = nil, want an error")
	}
	if err := svc.RefundAll(t.Context(), "p1"); err != nil {
		t.Fatalf("RefundAll(p1) = %v, want nil", err)
	}
	p, err := store.Get(t.Context(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Refunded != 1000 || len(gw.refunded) != 2 {
		t.Errorf("after refunds: Refunded = %d, gateway calls %v; want 1000 and [600 400]", p.Refunded, gw.refunded)
	}
}
