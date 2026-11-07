package config

import (
	"fmt"
	"log"
	"os"
	"time"

	

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGormDB(cfg *DBConfig, isDev bool) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, cfg.TimeZone,
	)

	// Cấu hình GORM log theo môi trường
	logLevel := logger.Warn
	if isDev {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  isDev,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to database: %w", err)
	}

	// Cấu hình Connection Pool (Database/sql engine ngầm)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve DB instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Kiểm tra kết nối thực tế
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("Cannot ping to database: %w", err)
	}

	return db, nil
}