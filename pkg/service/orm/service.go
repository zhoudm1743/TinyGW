package orm

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"time"
)

var db *gorm.DB

func NewDb() (*gorm.DB, error) {
	var err error
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("os.Getwd(): %v", err)
	}
	dbFile := path.Join(dir, "tinyGW.db")
	// 检查数据库文件是否存在，不存在生成sqlite3文件
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		if _, createErr := os.Create(dbFile); createErr != nil {
			return nil, fmt.Errorf("创建数据库文件失败: %v", createErr)
		}
	}
	l := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
		})
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger:                 l,
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open(): %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取sql.DB对象失败: %v", err)
	}
	// 在初始化后执行 PRAGMA 命令
	db.Exec("PRAGMA journal_mode = WAL;")   // 使用 Write-Ahead Logging
	db.Exec("PRAGMA synchronous = NORMAL;") // 降低同步频率
	db.Exec("PRAGMA cache_size = -10000;")  // 设置 10MB 缓存

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("连接有效性检查失败: %v", err)
	}
	return db, nil
}

func GetDb() *gorm.DB {
	return db
}

var Module = fx.Provide(
	NewDb,
)
