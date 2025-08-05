package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
	"time"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/http/middleware"
	"tinyGW/pkg/service/logger"
)

type Service struct {
	Gin    *gin.Engine
	Server *http.Server
}

func NewService(c *conf.Config) *Service {
	gin.SetMode(c.Server.Mode)
	eng := gin.New()
	eng.Use(middleware.Cors()).Use(logger.GinLogger(), logger.GinRecovery(true))
	// 设置静态资源
	eng.StaticFS("/static", http.Dir("./public/webroot/static"))
	//engine.StaticFS("/resource", http.Dir("./webroot/resource"))
	eng.StaticFile("/favicon.ico", "./public/webroot/favicon.ico")
	eng.StaticFile("/platform-config.json", "./public/webroot/platform-config.json")
	eng.GET("/", func(c *gin.Context) {
		c.File("./public/webroot/index.html")
	})
	eng.NoRoute(func(c *gin.Context) {
		//c.File("./public/webroot/index.html")
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})
	addr := fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      eng,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  600 * time.Second,
	}
	return &Service{
		Gin:    eng,
		Server: server,
	}
}

var Module = fx.Provide(
	NewService,
)
