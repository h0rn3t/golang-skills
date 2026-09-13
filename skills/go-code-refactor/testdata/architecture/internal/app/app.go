// Package app is the composition root: it constructs repositories, services
// and handlers in dependency order and adapts one module's root contract to
// another module's consumer-side interface. It holds no business rules.
package app

import (
	"context"
	"database/sql"
	"net/http"

	billingsvc "example.com/shop/internal/billing/services"
	"example.com/shop/internal/order"
	"example.com/shop/internal/order/handlers"
	"example.com/shop/internal/order/repositories"
	ordersvc "example.com/shop/internal/order/services"
	"example.com/shop/internal/platform/clock"
)

// App is the wired application.
type App struct {
	Orders  *ordersvc.Service
	Billing *billingsvc.Service
	mux     *http.ServeMux
}

// New wires the modules over one database handle.
func New(db *sql.DB) *App {
	orders := ordersvc.New(repositories.NewOrderStore(db), clock.System{})
	billing := billingsvc.New(orderReader{orders})
	mux := http.NewServeMux()
	handlers.Register(mux, orders)
	return &App{Orders: orders, Billing: billing, mux: mux}
}

// Handler is the process's HTTP surface.
func (a *App) Handler() http.Handler { return a.mux }

// orderReader adapts the order module's root contract to billing's consumer
// interface. The two Summary types are distinct named types on purpose; the
// mapping is the whole job of this adapter and it carries no policy.
type orderReader struct{ orders *ordersvc.Service }

func (r orderReader) Summary(ctx context.Context, id string) (billingsvc.OrderSummary, error) {
	s, err := r.orders.Summary(ctx, id)
	if err != nil {
		return billingsvc.OrderSummary{}, mapNotFound(err)
	}
	return billingsvc.OrderSummary{OrderID: s.ID, TotalCents: s.TotalCents}, nil
}

func mapNotFound(err error) error {
	if err == order.ErrNotFound {
		return billingsvc.ErrOrderUnknown
	}
	return err
}
