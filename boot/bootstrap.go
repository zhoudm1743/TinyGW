package boot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"TinyGW/core/logger"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"gorm.io/gorm"
)

func registerHooks(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			Db, err := db.DB()
			if err == nil {
				Db.Close()
			}
			return nil
		},
	})
}

func Run() error {
	app := fx.New(
		Module,
		fx.WithLogger(
			func(log *logrus.Logger) fxevent.Logger {
				return logger.NewFxLogger(log)
			},
		),
		fx.Invoke(registerHooks),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("fx app start failed: %w", err)
	}
	// 等待中断信号优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	stopCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("fx app stop failed: %w", err)
	}
	return nil
}
