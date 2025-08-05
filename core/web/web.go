package web

import (
	"TinyGW/core/config"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Gin    *gin.Engine
	Server *http.Server
}

// GinLogrusMiddleware 返回一个将 gin 日志输出到 logrus 的中间件
func GinLogrusMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		log.Infof("%s %s %d %s", c.Request.Method, c.Request.URL.Path, status, latency)
	}
}

func NewService(cfg *config.Config, log *logrus.Logger) *Service {
	gin.SetMode("release")
	eng := gin.New()
	eng.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"OPTIONS", "GET", "POST", "DELETE", "PUT"},
		MaxAge:       1 * time.Hour,
	}))
	eng.Use(GinLogrusMiddleware(log), gin.Recovery())
	// 静态资源
	eng.StaticFS("/static", http.Dir("./public/webroot/static"))
	eng.StaticFile("/favicon.ico", "./public/webroot/favicon.ico")
	eng.StaticFile("/platform-config.json", "./public/webroot/platform-config.json")
	eng.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	eng.GET("/", func(c *gin.Context) { c.File("./public/webroot/index.html") })
	eng.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "页面不存在"})
	})
	return &Service{
		Gin:    eng,
		Server: nil, // 由生命周期钩子负责创建和管理
	}
}
