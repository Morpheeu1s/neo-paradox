package config

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// --- 1. Định nghĩa Context Key an toàn ---

type contextKey string

const (
	CorrelationIDKey  contextKey = "correlation_id"
	CorrelationHeader string     = "X-Correlation-ID"
)

// InjectCorrelationID đưa correlation ID vào context
func InjectCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// GetCorrelationID lấy correlation ID từ context
func GetCorrelationID(ctx context.Context) string {
	if val, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return val
	}
	return ""
}

// --- 2. Custom Handler tự động gán Correlation ID ---

type CorrelationHandler struct {
	slog.Handler
}

// Handle intercept từng log record để chèn thuộc tính từ context
func (h *CorrelationHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := GetCorrelationID(ctx); id != "" {
		r.AddAttrs(slog.String("correlation_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *CorrelationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CorrelationHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *CorrelationHandler) WithGroup(name string) slog.Handler {
	return &CorrelationHandler{Handler: h.Handler.WithGroup(name)}
}

// --- 3. Khởi tạo Logger chuẩn Production ---

func InitProductionLogger(levelVar *slog.LevelVar) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     levelVar, // Cho phép thay đổi level động lúc runtime
		AddSource: true,     // Hiện file và số dòng code
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Chuẩn hóa tên trường timestamp theo chuẩn OpenTelemetry / JSON schema
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339Nano))
			}

			// Rút gọn đường dẫn file source: chỉ lấy "folder/file.go:line" thay vì cả đường dẫn tuyệt đối
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					dir := filepath.Base(filepath.Dir(source.File))
					file := filepath.Base(source.File)
					return slog.String("source", fmt.Sprintf("%s/%s:%d", dir, file, source.Line))
				}
			}

			return a
		},
	}

	// Base JSON Handler ghi ra stdout
	baseHandler := slog.NewJSONHandler(os.Stdout, opts)

	// Bọc bằng CorrelationHandler
	correlationHandler := &CorrelationHandler{Handler: baseHandler}

	// Gắn metadata cấp tiến trình (Service Name, Version, Môi trường)
	logger := slog.New(correlationHandler).With(
		slog.String("service.name", "order-service"),
		slog.String("service.version", "1.4.0"),
		slog.String("environment", "production"),
	)

	// Đặt làm Logger mặc định
	slog.SetDefault(logger)

	return logger
}

// --- 4. HTTP Middleware trích xuất và gắn ID ---

func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get(CorrelationHeader)
		if corrID == "" {
			corrID = uuid.NewString()
		}

		// Gắn lại vào header response để client hoặc microservice gọi tiếp có thể trace
		w.Header().Set(CorrelationHeader, corrID)

		// Lưu vào context
		ctx := InjectCorrelationID(r.Context(), corrID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
