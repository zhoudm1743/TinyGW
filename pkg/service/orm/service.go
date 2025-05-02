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
		Logger: l,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open(): %v", err)
	}
	return db, nil
}

func GetDb() *gorm.DB {
	return db
}

func CloseDb() {
	sqlDB, _ := db.DB()
	err := sqlDB.Close()
	if err != nil {
		fmt.Println("关闭数据库连接失败: ", err)
		return
	}
}

var Module = fx.Provide(
	NewDb,
)
