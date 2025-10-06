package core

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gostratum/core/pkg/logx"
)

// RunHTTP bootstraps the application and powers the HTTP lifecycle until a
// termination signal is received.
func RunHTTP(opts BuildOptions, factory HTTPHandlerFactory) error {
	if factory == nil {
		return errors.New("core: http handler factory is nil")
	}

	app, err := Bootstrap(opts)
	if err != nil {
		return err
	}

	handler := factory(Deps{Config: app.Cfg})
	if handler == nil {
		return errors.New("core: http handler is nil")
	}

	server := &http.Server{
		Addr:              app.Cfg.Server.Addr,
		Handler:           handler,
		ReadHeaderTimeout: app.Cfg.Server.ReadHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if app.Log != nil {
			app.Log.Info("starting http server", logx.Fields{"addr": server.Addr})
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if app.Log != nil {
				app.Log.Error("http server failed", err, logx.Fields{"addr": server.Addr})
			}
			return err
		}

		return nil
	})

	group.Go(func() error {
		<-groupCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if app.Log != nil {
			app.Log.Info("shutting down http server", logx.Fields{"addr": server.Addr})
		}

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if app.Log != nil {
				app.Log.Error("http server shutdown failed", err, logx.Fields{"addr": server.Addr})
			}
			return err
		}

		if app.Log != nil {
			app.Log.Info("http server stopped", logx.Fields{"addr": server.Addr})
		}

		return nil
	})

	return group.Wait()
}
