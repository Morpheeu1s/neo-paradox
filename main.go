package main

import (
	"context"
	"fmt"
	"golearn/internal/platform/config"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Cannot load data environemt", slog.String("error", err.Error()))
		os.Exit(1)
	}
	initLogger(cfg)

	slog.Info("Application environment loaded successfully!")
	ctx := context.Background()
	slog.InfoContext(ctx, "Init success")

	db, err := gorm.Open(postgres.Open(createConnectionString(cfg)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while connecting to database: %s", err.Error()))
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while initiating database instance: %s", err.Error()))
		os.Exit(1)
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		slog.ErrorContext(ctx, "Ping DB failed: %w", err)
		os.Exit(1)
	}
	r := gin.Default()
	slog.InfoContext(ctx, "Server run at: http://localhost:8080/")
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Lỗi chạy server: %v", err.Error()))
	}
	
}

func createConnectionString(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)
}

func initLogger(cfg *config.Config) {
	// Xác định Log Level từ config (mặc định INFO nếu không truyền)
	var level slog.Level
	switch cfg.LogLevel {
	case "DEBUG", "debug":
		level = slog.LevelDebug
	case "WARN", "warn":
		level = slog.LevelWarn
	case "ERROR", "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler

	// Môi trường production ưu tiên JSON để đẩy vào ELK/Loki/Datadog
	if cfg.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		// Môi trường local/dev dùng tint Handler: có màu sắc, format gọn gàng
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05", // Định dạng giờ:phút:giây gọn gàng
			AddSource:  true,       // Tự động rút gọn thành main.go:23 thay vì đường dẫn tuyệt đối dài
		})
	}

	// Gắn thông tin chung của service vào mọi log
	logger := slog.New(handler).With(
		slog.String("service", "golearn-api"),
		slog.String("env", cfg.Environment),
	)

	// Đặt làm Logger mặc định cho toàn bộ package/ứng dụng
	slog.SetDefault(logger)
}
