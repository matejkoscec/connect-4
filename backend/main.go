package main

import (
	"backend/cache"
	"backend/config"
	"backend/generated/sqlc"
	"backend/handlers"
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	cfg, err := config.LoadConfig(e.Logger)
	if err != nil {
		return err
	}
	fmt.Print(cfg.PrettyString())

	e.Logger.SetLevel(cfg.App.Logger.Level.Lvl)
	e.Validator = &RequestValidator{validator.New()}

	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dbpool, err := pgxpool.New(ctx, cfg.App.DB.URL)
	if err != nil {
		return err
	}
	defer dbpool.Close()

	queries := sqlc.New(dbpool)
	gameCache := cache.NewDefaultCache(queries, dbpool)
	h := &handlers.Handler{
		DB:        queries,
		Config:    *cfg,
		Conn:      dbpool,
		GameCache: gameCache,
		BaseCtx:   ctx,
	}

	handlers.ConfigureRoutes(h, e)
	go gameCache.RunMatchmaking(ctx)

	go func() {
		port := fmt.Sprintf(":%d", cfg.App.Port)
		if err := e.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	sdCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err = e.Shutdown(sdCtx); err != nil {
		return err
	}

	return nil
}
