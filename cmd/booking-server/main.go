package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"spacetrouble/internal/pkg/booking"
	"spacetrouble/internal/pkg/config"
	"spacetrouble/internal/pkg/data/postgres"
	"spacetrouble/internal/pkg/health"
	"spacetrouble/internal/pkg/spacex"
	"spacetrouble/pkg/apiutils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		signalTrapped := <-c
		fmt.Println(signalTrapped)
		cancel()
	}()

	cfg := config.NewConfig()

	if err := run(ctx, cfg); err != nil {
		fmt.Println(err)
	} else {
	}
}

type serviceContainer struct {
	bookSrv booking.BookingService
}

func run(ctx context.Context, cfg *config.Config) (err error) {
	db, err := pgxpool.Connect(ctx, cfg.DSN())
	if err != nil {
		return err
	}
	if err := db.Ping(ctx); err != nil {
		return err
	}
	defer db.Close()

	store := postgres.NewStore(db)
	spaceXClient := spacex.NewSpaceXClient(cfg.SpaceXUrl)
	srvC := serviceContainer{
		bookSrv: booking.NewBookingService(store, spaceXClient),
	}

	router := setupRouter(ctx, srvC)

	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		WriteTimeout: cfg.ServerWriteTimeout,
		ReadTimeout:  cfg.ServerReadTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
		Handler:      router,
	}

	srvErrC := make(chan error, 1)
	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			srvErrC <- err
		}
	}()

	select {
	case <-ctx.Done():
	case srvErr := <-srvErrC:
		err = srvErr
	}

	if err != nil {
		return
	}

	ctxClose, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctxClose); err != nil {
		return
	}

	if err == http.ErrServerClosed {
		err = nil
	}
	return
}

func setupRouter(ctx context.Context, srvC serviceContainer) http.Handler {
	router := http.NewServeMux()

	versionPrefix := "/v1"

	router.HandleFunc(versionPrefix+"/health", health.HealthGet())

	bookingHandler := apiutils.AllowedMethods(
		apiutils.AllowedContentTypes(booking.BookingHandler(srvC.bookSrv), "application/json"),
		"POST", "GET",
	)
	router.HandleFunc(versionPrefix+"/bookings", bookingHandler)

	return router
}
