package database

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"go.uber.org/fx"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGorm(db *gorm.DB) (*gorm.DB, error) {
	var err error
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前工作目录失败: %v", err)
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
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
		})
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger:                 l,
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	return db, err
}

var Module = fx.Provide(
	NewGorm,
)
