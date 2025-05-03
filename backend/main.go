package main

import (
	"backend/handlers"
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	e := echo.New()
	if err := run(e); err != nil {
		e.Logger.Fatal(err)
	}
}

func run(e *echo.Echo) error {
	e.Logger.SetLevel(log.INFO)

	bgContext := context.Background()

	dbpool, err := pgxpool.New(bgContext, os.Getenv("DB_URL"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer dbpool.Close()

	handlers.ConfigureRoutes(e, dbpool)

	e.Validator = &RequestValidator{validator.New()}

	ctx, stop := signal.NotifyContext(bgContext, os.Interrupt)
	defer stop()
	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(bgContext, 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
