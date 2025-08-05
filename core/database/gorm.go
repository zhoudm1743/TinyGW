package database

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LogrusGormLogger 适配器
// 支持 GORM 日志输出到 logrus
// Level: info/warn/error
// Colorful: 由 logrus 控制

type LogrusGormLogger struct {
	Log      *logrus.Logger
	LogLevel logger.LogLevel
	Colorful bool
}

func (l *LogrusGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &LogrusGormLogger{
		Log:      l.Log,
		LogLevel: level,
		Colorful: l.Colorful,
	}
}

func (l *LogrusGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Log.Infof(msg, data...)
	}
}

func (l *LogrusGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Log.Warnf(msg, data...)
	}
}

func (l *LogrusGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Log.Errorf(msg, data...)
	}
}

func (l *LogrusGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	msg := fmt.Sprintf("[%.3f毫秒] [行数:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	if err != nil && l.LogLevel >= logger.Error {
		l.Log.WithError(err).Error("SQL执行出错: " + msg)
	} else if elapsed > 200*time.Millisecond && l.LogLevel >= logger.Warn {
		l.Log.Warn("SQL慢查询: " + msg)
	} else if l.LogLevel >= logger.Info {
		l.Log.Info("SQL执行: " + msg)
	}
}

func NewGorm(log *logrus.Logger) (*gorm.DB, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前工作目录失败: %v", err)
	}
	dbFile := path.Join(dir, "tinyGW.db")
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		if _, createErr := os.Create(dbFile); createErr != nil {
			return nil, fmt.Errorf("创建数据库文件失败: %v", createErr)
		}
	}
	gormLogger := &LogrusGormLogger{
		Log:      log,
		LogLevel: logger.Warn,
		Colorful: true,
	}
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger:                 gormLogger,
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	return db, nil
}

var Module = fx.Provide(
	NewGorm,
)
