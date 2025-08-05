package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"TinyGW/core/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func autoListen(basePort string, maxTry int, log *logrus.Logger) (net.Listener, string, error) {
	portNum, _ := strconv.Atoi(basePort)
	for i := 0; i < maxTry; i++ {
		addr := ":" + strconv.Itoa(portNum)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			if i > 0 {
				log.Warnf("端口 %s 被占用，自动切换到可用端口 %d", basePort, portNum)
			}
			return ln, strconv.Itoa(portNum), nil
		}
		portNum++
	}
	return nil, "", errors.New("无法找到可用端口")
}

func registerWebHooks(lc fx.Lifecycle, svc *Service, log *logrus.Logger, cfg *config.Config) {
	basePort := cfg.Server.Port
	if basePort == "" {
		basePort = "8000"
	}
	ln, finalPort, err := autoListen(basePort, 10, log)
	if err != nil {
		panic(err)
	}
	server := &http.Server{
		Addr:         ":" + finalPort,
		Handler:      svc.Gin,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  600 * time.Second,
	}
	svc.Server = server
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Infof("HTTP server started on port %s", finalPort)
			go func() {
				if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Errorf("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewService),
	fx.Invoke(registerWebHooks),
)
