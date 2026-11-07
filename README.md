# GoLearn API

Dự án mẫu REST API backend viết bằng ngôn ngữ **Go (Golang)**, tích hợp framework **Gin**, ORM **GORM (PostgreSQL)**, hệ thống ghi log có cấu trúc chuẩn **`log/slog`** với định dạng màu sắc đẹp mắt khi dev (`tint`), cùng khả năng quản lý cấu hình tập trung qua biến môi trường.

---

## 🚀 Tính năng nổi bật

- **Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin) tốc độ cao, hỗ trợ middleware linh hoạt.
- **Database & ORM:** [GORM](https://gorm.io/) kết nối PostgreSQL với cấu hình Connection Pooling tối ưu (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `ConnMaxIdleTime`).
- **Structured Logging (`slog`):**
  - Môi trường **Development**: Sử dụng [`github.com/lmittmann/tint`](https://github.com/lmittmann/tint) hiển thị log có màu sắc, rút gọn đường dẫn file nguồn và định dạng thời gian gọn gàng (`15:04:05`).
  - Môi trường **Production**: Tự động chuyển sang `slog.NewJSONHandler` chuẩn cho hệ thống thu thập log (ELK, Loki, Datadog).
  - Hỗ trợ **Correlation ID / Request Tracing** qua HTTP Middleware và Context.
- **Quản lý cấu hình:** Sử dụng [`caarlos0/env`](https://github.com/caarlos0/env) và [`godotenv`](https://github.com/joho/godotenv) tự động bind biến môi trường vào struct Go với giá trị mặc định (`envDefault`) và ràng buộc bắt buộc (`required`).

---

## 📁 Cấu trúc thư mục

```text
golearn/
├── internal/
│   └── platform/
│       └── config/
│           ├── connect.go     # Hàm khởi tạo kết nối GORM PostgreSQL & Connection Pool
│           ├── dbconfig.go    # Định nghĩa struct Config/DBConfig và nạp file .env
│           └── logconfig.go   # Tiện ích Correlation ID & Logger chuẩn Production
├── .env                       # Biến môi trường local (chứa secret, không commit lên git)
├── .env.example               # File mẫu cấu hình biến môi trường
├── go.mod                     # Go module & dependencies
├── go.sum                     # Checksums dependencies
├── main.go                    # Entrypoint ứng dụng: nạp config, logger, DB, Gin router
└── README.md                  # Tài liệu hướng dẫn dự án
```

---

## 🛠️ Yêu cầu môi trường

- **Go:** Phiên bản `>= 1.22` (dự án đang sử dụng Go `1.26+`).
- **PostgreSQL:** Phiên bản `>= 14`.

---

## ⚙️ Cấu hình biến môi trường (`.env`)

Tạo file `.env` ở thư mục gốc của dự án với nội dung mẫu như sau:

```env
# Server
PORT=8080
APP_ENV=development
LOG_LEVEL=INFO
ENV_MODE=development

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=paradox
DB_PASSWORD=paradox123
DB_NAME=paradox
DB_SSL_MODE=disable
DB_TIMEZONE=Asia/Ho_Chi_Minh

# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=30m
DB_CONN_MAX_IDLE_TIME=15m
```

### Chi tiết các biến cấu hình

| Biến môi trường | Mặc định | Bắt buộc | Mô tả |
| :--- | :--- | :---: | :--- |
| `PORT` | `8080` | Không | Cổng HTTP server lắng nghe |
| `APP_ENV` / `ENV_MODE` | `development` | Không | Môi trường chạy (`development` hoặc `production`) |
| `LOG_LEVEL` | `INFO` | Không | Mức độ log (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `DB_HOST` | - | **Có** | Địa chỉ máy chủ PostgreSQL |
| `DB_PORT` | `5432` | Không | Cổng kết nối PostgreSQL |
| `DB_USER` | - | **Có** | Tên người dùng database |
| `DB_PASSWORD` | - | **Có** | Mật khẩu database |
| `DB_NAME` | - | **Có** | Tên database cần kết nối |
| `DB_SSL_MODE` | `disable` | Không | Chế độ SSL (`disable`, `require`, `verify-full`,...) |
| `DB_TIMEZONE` | `UTC` | Không | Múi giờ của PostgreSQL session |
| `DB_MAX_OPEN_CONNS` | `25` | Không | Số lượng connection mở tối đa vào DB |
| `DB_MAX_IDLE_CONNS` | `10` | Không | Số lượng connection rảnh tối đa được giữ lại |
| `DB_CONN_MAX_LIFETIME` | `30m` | Không | Thời gian sống tối đa của 1 connection |
| `DB_CONN_MAX_IDLE_TIME` | `15m` | Không | Thời gian rảnh tối đa trước khi connection bị đóng |

---

## 🏃 Hướng dẫn cài đặt & khởi chạy

### 1. Tải dependencies

```bash
go mod tidy
```

### 2. Chuẩn bị cơ sở dữ liệu PostgreSQL

Đảm bảo dịch vụ PostgreSQL đang chạy và cơ sở dữ liệu đã được tạo:

```sql
CREATE DATABASE paradox;
CREATE USER paradox WITH ENCRYPTED PASSWORD 'paradox123';
GRANT ALL PRIVILEGES ON DATABASE paradox TO paradox;
```

### 3. Chạy ứng dụng ở môi trường Local / Development

```bash
go run main.go
```

Khi chạy thành công, bạn sẽ thấy log màu sắc từ `tint` hiển thị:
```text
20:31:56 INF golearn\main.go:25 Application environment loaded successfully! service=golearn-api env=development
20:31:56 INF golearn\main.go:27 Init success service=golearn-api env=development
20:31:56 INF golearn\main.go:50 Server run at: http://localhost:8080/ service=golearn-api env=development
[GIN-debug] Listening and serving HTTP on :8080
```

### 4. Build binary cho Production

```bash
# Build binary
go build -o golearn-api.exe main.go

# Chạy với biến môi trường Production
ENV_MODE=production ./golearn-api.exe
```

Ở chế độ `production`, log sẽ tự động xuất ra định dạng JSON 1 dòng chuẩn hoá cho hệ thống log collector.

---

## 📜 Giấy phép

Dự án phát triển cho mục đích học tập và xây dựng REST API bằng Go.
