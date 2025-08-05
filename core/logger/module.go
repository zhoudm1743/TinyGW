package logger

import (
	"os"
	"strings"

	"TinyGW/core/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

// NewLogger 创建带颜色的 logrus.Logger，支持 level/format 配置
func NewLogger(cfg *config.Config) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	// 设置格式
	if strings.ToLower(cfg.Logger.Format) == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置日志级别
	switch strings.ToLower(cfg.Logger.Level) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}

var Module = fx.Provide(NewLogger)
