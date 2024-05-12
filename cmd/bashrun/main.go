package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	"golang.org/x/sync/semaphore"

	"github.com/PoorMercymain/bashrun/internal/bashrun/config"
	"github.com/PoorMercymain/bashrun/internal/bashrun/handler"
	"github.com/PoorMercymain/bashrun/internal/bashrun/repository"
	"github.com/PoorMercymain/bashrun/internal/bashrun/service"
	"github.com/PoorMercymain/bashrun/pkg/logger"
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

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(cfg.MaxConcurrentCommands)
	commandContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := repository.New(pg)
	s := service.New(commandContext, r, sem, &wg)
	h := handler.New(s)

	mux := http.NewServeMux()

	mux.Handle("GET /ping", http.HandlerFunc(h.Ping))
	mux.Handle("POST /commands", http.HandlerFunc(h.CreateCommand))
	mux.Handle("GET /commands", http.HandlerFunc(h.ListCommands))
	mux.Handle("GET /commands/stop/{command_id}", http.HandlerFunc(h.StopCommand))
	mux.Handle("GET /commands/{command_id}", http.HandlerFunc(h.ReadCommand))
	mux.Handle("GET /commands/output/{command_id}", http.HandlerFunc(h.ReadOutput))

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

	ctx, cancel = context.WithTimeout(commandContext, time.Second*5)
	defer cancel()

	wgChan := make(chan struct{})
	go func() {
		wg.Wait()
		wgChan <- struct{}{}
	}()

	select {
	case <-wgChan:
	case <-ctx.Done():
		logger.Logger().Errorln("some of the commands were interrupted")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	poolCloseCh := make(chan struct{})
	go func() {
		pool.Close()
		poolCloseCh <- struct{}{}
	}()

	select {
	case <-poolCloseCh:
	case <-ctx.Done():
		logger.Logger().Errorln("postgres pool forced to close")
	}
}
