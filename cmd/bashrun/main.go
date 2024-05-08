package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PoorMercymain/bashrun/internal/bashrun/config"
	"github.com/PoorMercymain/bashrun/internal/bashrun/handler"
	"github.com/PoorMercymain/bashrun/internal/bashrun/repository"
	"github.com/PoorMercymain/bashrun/internal/bashrun/service"
	"github.com/PoorMercymain/bashrun/pkg/logger"
	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Logger().Fatalln("Failed to parse env: %v", err)
	}

	logger.SetLogFile("logs/" + cfg.LogFilePath)
	m, err := migrate.New("file://"+cfg.MigrationsPath, cfg.DSN())
	if err != nil {
		logger.Logger().Fatalln(err.Error())
	}

	err = repository.ApplyMigrations(m)
	if err != nil {
		logger.Logger().Fatalln(err.Error())
	}

	logger.Logger().Infoln("Migrations applied successfully")

	pool, err := repository.GetPgxPool(cfg.DSN())
	if err != nil {
		logger.Logger().Fatalln(err)
	}

	logger.Logger().Infoln("Postgres connection pool created")

	pg := repository.NewPostgres(pool)

	r := repository.New(pg)
	s := service.New(r)
	h := handler.New(s)

	mux := http.NewServeMux()

	mux.Handle("GET /ping", http.HandlerFunc(h.Ping))

	server := &http.Server{
		Addr:     fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort),
		ErrorLog: log.New(logger.Logger(), "", 0),
		Handler:  mux,
	}

	go func() {
		logger.Logger().Infoln("Server started, listening on port", cfg.ServicePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger().Fatalln("ListenAndServe failed:", err.Error())
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = server.Shutdown(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Logger().Errorln("server forced to shutdown")
	} else if err != nil {
		logger.Logger().Errorln("error while shutting down server:", err.Error())
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	poolCloseCh := make(chan struct{})
	go func () {
		pool.Close()
		poolCloseCh<-struct{}{}
	}()

	select {
	case <-poolCloseCh:
	case <-ctx.Done():
		logger.Logger().Errorln("postgres pool forced to close")
	}
}