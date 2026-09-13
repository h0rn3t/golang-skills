// Command server is the process entry point: it builds the application and
// serves it. Business rules live in the modules, wiring in internal/app.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"example.com/shop/internal/app"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	dsn := "postgres://localhost/shop?sslmode=disable"
	if len(args) > 0 {
		dsn = args[0]
	}
	db, err := sql.Open("postgres", dsn) // the driver is registered by the production build, not this fixture
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app.New(db).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
