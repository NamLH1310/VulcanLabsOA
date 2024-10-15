package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/namlh/vulcanLabsOA/config"
	"github.com/namlh/vulcanLabsOA/controller"
	"github.com/namlh/vulcanLabsOA/logging"
	"github.com/namlh/vulcanLabsOA/manager"
	"github.com/namlh/vulcanLabsOA/middleware"
	"github.com/namlh/vulcanLabsOA/util/fmtutil"
)

func Run(ctx context.Context, getEnv func(key string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := config.Read(getEnv)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	logger, err := logging.NewLogger(&cfg.Logger, cfg.Env)
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}

	// managers
	groupManager := manager.NewGroupManager(cfg.Groups)
	roomManager := manager.NewRoomManager(
		logger,
		&cfg.Room,
		groupManager,
	)

	// controllers
	groupController := controller.NewGroupController(logger, groupManager)
	roomController := controller.NewRoomController(logger, roomManager)

	srv := NewServer(
		logger,
		roomController,
		groupController,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		Handler: srv,
	}

	go func() {
		logger.InfoContext(ctx, "listening and serve", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmtutil.Eprintf("error listening and serving: %s\n", err)
		}
	}()

	// graceful shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmtutil.Eprintf("error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func addRoutes(
	logger *slog.Logger,
	mux *http.ServeMux,
	roomController *controller.RoomController,
	groupController *controller.GroupController,
) {
	const apiPathPrefix = "/api/v1"

	type HandlerConfig struct {
		method  string
		path    string
		handler http.HandlerFunc
	}

	handlerConfigs := []HandlerConfig{
		{"GET", "/health", controller.HealthCheck(logger)},
		{"GET", "/groups", groupController.ListGroupIDs},

		{"GET", "/available-seats", roomController.ListAvailableSeats},
		{"POST", "/seats/reservation", roomController.ReserveSeats},
		{"POST", "/seats/cancellation", roomController.CancelSeats},
	}

	for _, cfg := range handlerConfigs {
		if len(cfg.path) == 0 {
			fmtutil.Eprintf("invalid handler path")
			os.Exit(1)
		}
		mux.Handle(cfg.method+" "+path.Join(apiPathPrefix, cfg.path), cfg.handler)
	}
}

func NewServer(
	logger *slog.Logger,
	roomController *controller.RoomController,
	groupController *controller.GroupController,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		logger,
		mux,
		roomController,
		groupController,
	)

	var httpHandler http.Handler = mux

	httpHandler = middleware.PanicRecover(logger, httpHandler)
	httpHandler = middleware.Logging(logger, httpHandler)

	return httpHandler
}
