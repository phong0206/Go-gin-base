# Go Service Code Base Structure

Tài liệu này chốt hướng thiết kế ban đầu cho service Go theo mô hình modular monolith, có thể mở rộng dần sang Redis, Kafka và background worker.

## Tech stack đã chốt

### Core

| Thành phần | Lựa chọn | Mục đích |
| --- | --- | --- |
| Language | Go | Ngôn ngữ chính |
| HTTP framework | `github.com/gin-gonic/gin` | Router, middleware, binding request và HTTP response |
| Database | PostgreSQL | Database chính cho user, order, ticket |
| SQL access | `github.com/jmoiron/sqlx` | Mở rộng từ `database/sql`, hỗ trợ `Get`, `Select`, named query và scan vào struct |
| PostgreSQL driver | `github.com/jackc/pgx/v5/stdlib` | Driver PostgreSQL cho `database/sql`/`sqlx` |
| Migration | Tự viết DDL | Team chủ động quản lý schema; không dùng auto migration từ entity |
| Configuration | Environment variables | Tách config khỏi source code |
| Logging | `log/slog` | Logger chuẩn của Go, structured logging |
| Testing | `testing` + `httptest` | Unit test và HTTP handler test |

### Authentication

| Thành phần | Lựa chọn | Mục đích |
| --- | --- | --- |
| Password hashing | `golang.org/x/crypto/bcrypt` | Hash và verify password |
| Token | `github.com/golang-jwt/jwt/v5` | Access token/refresh token nếu dùng JWT |
| Validation | Validator tích hợp với Gin binding | Validate request DTO |

### Thành phần mở rộng theo phase

| Thành phần | Lựa chọn | Dùng khi |
| --- | --- | --- |
| Cache | `github.com/redis/go-redis/v9` | Cần cache, rate limit hoặc lưu session/token |
| Queue | `github.com/hibiken/asynq` | Cần background job dựa trên Redis |
| Event streaming | `github.com/segmentio/kafka-go` hoặc `franz-go` | Cần Kafka và event-driven processing |
| API documentation | `swaggo/swag` hoặc OpenAPI | Cần Swagger/OpenAPI |
| Local hot reload | `air-verse/air` | Chỉ dùng trong development |
| Lint | `golangci-lint` | Static analysis và kiểm tra coding convention |

Không đưa Redis, Kafka hoặc queue vào phase đầu nếu auth/user chưa ổn định. Trước hết nên hoàn thiện boundary và interface để thêm các thành phần này mà không phải sửa business logic.

## Vì sao dùng sqlx thay vì ORM?

`sqlx` không phải ORM. Nó là một lớp tiện ích trên `database/sql`:

- Viết SQL trực tiếp, kiểm soát rõ query và index.
- Scan kết quả vào struct bằng `db` tag.
- Hỗ trợ `Get`, `Select`, `NamedExec`, transaction.
- Không tự sinh migration từ entity.
- Không có lazy loading, relationship magic hoặc query behavior khó đoán như một số ORM.

Ví dụ convention:

```go
type User struct {
    ID    uuid.UUID `db:"id" json:"id"`
    Email string    `db:"email" json:"email"`
}
```

Repository sẽ sở hữu SQL và database mapping. Application service không được biết SQL cụ thể.

```text
handler -> application service -> repository interface -> postgres repository -> sqlx -> PostgreSQL
```

## Migration

Migration được viết thủ công bằng DDL. Có thể dùng một migration runner bên ngoài, nhưng nội dung schema vẫn do developer kiểm soát.

```text
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_orders.up.sql
└── 000002_create_orders.down.sql
```

Quy ước:

- Mỗi thay đổi schema tạo một version mới.
- Không sửa migration đã chạy trên môi trường shared/production.
- Migration `up` dùng để apply thay đổi.
- Migration `down` dùng để rollback trong local/test khi phù hợp.
- Luôn review index, constraint, foreign key và backward compatibility.
- Không dùng `AutoMigrate` trong production.

## Cấu trúc tổng thể

```text
root/
├── cmd/
│   ├── root.go
│   ├── http.go
│   └── kafka.go
│
├── docker/
│   ├── api.Dockerfile
│   └── db/
│       └── initdb/
│           └── postgres-init.sh
│
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_orders.up.sql
│   └── 000002_create_orders.down.sql
│
├── src/
│   ├── bootstrap/
│   │   ├── http.go
│   │   └── kafka.go
│   │
│   ├── applications/
│   │   ├── commands/
│   │   ├── queries/
│   │   ├── dto/
│   │   ├── listeners/
│   │   └── services/
│   │
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user.go
│   │   │   └── order.go
│   │   ├── valueobjects/
│   │   │   ├── email.go
│   │   │   ├── money.go
│   │   │   └── timestamp.go
│   │   ├── repositories/
│   │   │   ├── user_repository.go
│   │   │   └── order_repository.go
│   │   ├── events/
│   │   │   ├── order_created.go
│   │   │   └── user_registered.go
│   │   └── errors/
│   │       ├── bad_request.go
│   │       └── internal_server_error.go
│   │
│   ├── infrastructure/
│   │   ├── config/
│   │   ├── constants/
│   │   ├── persistence/
│   │   │   ├── database.go
│   │   │   └── postgres/
│   │   │       ├── user_repository.go
│   │   │       └── order_repository.go
│   │   ├── messaging/
│   │   │   ├── kafka/
│   │   │   └── rabbitmq/
│   │   ├── external/
│   │   │   ├── aws/
│   │   │   └── ...
│   │   └── utils/
│   │
│   └── interfaces/
│       ├── rest/
│       │   ├── handlers/
│       │   ├── middleware/
│       │   └── router.go
│       ├── grpc/
│       └── kafka/
│
├── tests/
│   ├── unit/
│   ├── integration/
│   │   ├── database/
│   │   └── interfaces/
│   └── mocks/
│
├── storage/
│   └── logs/
│
├── .env.example
├── .gitignore
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Giải thích từng thư mục và file

### Root files

| File | Công dụng |
| --- | --- |
| `main.go` | Entry point tối thiểu của application. Đọc command/root và gọi bootstrap phù hợp. Không chứa business logic. |
| `go.mod` | Khai báo module path, Go version và dependency trực tiếp. Tương đương một phần vai trò của `package.json`. |
| `go.sum` | Lưu checksum của dependency. File này được Go tự quản lý và nên commit vào Git. |
| `Makefile` | Chuẩn hóa các lệnh `run`, `test`, `lint`, `build`, migration và generate cho team. |
| `Dockerfile` | Build image production của service. Nếu dùng nhiều image riêng thì có thể dùng `docker/api.Dockerfile`. |
| `.env.example` | Danh sách các biến môi trường cần có, không chứa secret thật. |
| `.gitignore` | Loại trừ `.env`, binary, log, coverage và file local khỏi Git. |
| `README.md` | Hướng dẫn chạy project, setup database, test, build và các quy ước cơ bản. |

### `cmd/`

Chứa command entrypoint. Mỗi command có thể build/deploy thành một process riêng.

| File | Công dụng |
| --- | --- |
| `cmd/root.go` | Root command, khai báo flag/config dùng chung và các subcommand. |
| `cmd/http.go` | Command khởi động HTTP API bằng Gin. |
| `cmd/kafka.go` | Command khởi động Kafka consumer/producer process nếu hệ thống cần Kafka. |

HTTP API và Kafka worker nên là hai process/container riêng để scale độc lập. Không bắt buộc start cả hai trong cùng một process.

### `docker/`

Chứa các file phục vụ build và khởi tạo môi trường Docker.

| Path | Công dụng |
| --- | --- |
| `docker/api.Dockerfile` | Multi-stage build cho API binary. Dùng khi muốn tách Dockerfile khỏi root. |
| `docker/db/initdb/postgres-init.sh` | Script khởi tạo PostgreSQL local, tạo extension/database/user nếu cần. Không thay thế migration schema. |

`postgres-init.sh` chỉ nên xử lý initialization của container. Bảng và index vẫn do migration quản lý.

### `migrations/`

Chứa DDL versioned do developer tự viết.

| File | Công dụng |
| --- | --- |
| `*_create_users.up.sql` | Tạo hoặc thay đổi schema theo chiều tiến. |
| `*_create_users.down.sql` | Rollback migration trong local/test khi cần. |

Migration không đặt trong `src` vì đây là database lifecycle, không phải runtime business code.

### `src/bootstrap/`

Composition root của từng transport/process. Nơi wire dependency từ ngoài vào trong.

| File | Công dụng |
| --- | --- |
| `bootstrap/http.go` | Load config, mở PostgreSQL, tạo repository/service/handler, đăng ký Gin route và tạo HTTP server. |
| `bootstrap/kafka.go` | Tạo Kafka client, consumer, listener và dependency cần cho worker. |

Bootstrap được phép biết tất cả layer. Các layer bên trong không được import ngược bootstrap.

### `src/applications/`

Chứa use case của hệ thống. Đây là application layer, điều phối domain và repository.

| Folder | Công dụng |
| --- | --- |
| `commands/` | Use case làm thay đổi state: register user, update user, create order. |
| `queries/` | Use case đọc dữ liệu: get current user, list orders, search ticket. |
| `dto/` | Input/output model của application layer, không nhất thiết trùng entity database. |
| `listeners/` | Xử lý event/message từ Kafka hoặc queue rồi gọi application use case. |
| `services/` | Logic orchestration dùng chung giữa command/query khi chưa cần tách thành nhiều package. |

Không đặt Gin `Context`, SQL query hoặc Kafka implementation trực tiếp trong application service.

### `src/domain/`

Chứa business rule thuần, hạn chế dependency vào Gin, PostgreSQL, Kafka hay Redis.

| Folder | Công dụng |
| --- | --- |
| `entities/` | Object có identity và lifecycle, ví dụ `User`, `Order`. |
| `valueobjects/` | Giá trị có validation/invariant riêng, ví dụ `Email`, `Money`, `Timestamp`. |
| `repositories/` | Interface repository mà application cần. Implementation nằm ở infrastructure. |
| `events/` | Domain event, ví dụ `UserRegistered`, `OrderCreated`. |
| `errors/` | Domain/application error có thể map sang HTTP hoặc message error. |

Ví dụ dependency direction:

```text
domain <- applications <- interfaces
domain <- infrastructure
```

`domain` không được import `interfaces` hoặc `infrastructure`.

### `src/infrastructure/`

Chứa implementation phụ thuộc external technology.

| Folder/File | Công dụng |
| --- | --- |
| `config/` | Parse environment variables thành typed config. |
| `constants/` | Constant dùng xuyên infrastructure; constant business nên đặt gần domain. |
| `persistence/database.go` | Mở, cấu hình và đóng `*sql.DB`; thiết lập pool, timeout, ping. |
| `persistence/postgres/` | Implementation repository dùng `sqlx` và PostgreSQL. |
| `messaging/kafka/` | Kafka producer, consumer và adapter. |
| `messaging/rabbitmq/` | RabbitMQ adapter nếu sau này cần; không tạo nếu chưa dùng. |
| `external/aws/` | Adapter gọi S3, SES, SQS hoặc AWS service khác. |
| `utils/` | Chỉ chứa helper infrastructure thực sự dùng chung. Tránh biến thành nơi chứa code không có chủ sở hữu. |

### `src/interfaces/`

Chứa adapter nhận request từ bên ngoài và chuyển vào application layer.

| Folder/File | Công dụng |
| --- | --- |
| `rest/router.go` | Tạo Gin engine, đăng ký route group và middleware. |
| `rest/handlers/` | Parse request, bind/validate DTO, gọi use case và format HTTP response. |
| `rest/middleware/` | Authentication, authorization, request ID, recovery, logging, CORS, rate limit. |
| `grpc/` | gRPC server adapter nếu cần public/internal gRPC API. |
| `kafka/` | Adapter cho message transport; nhận message rồi gọi application listener. |

Handler không truy cập database trực tiếp. Handler chỉ nên làm transport concern.

### `tests/`

| Folder | Công dụng |
| --- | --- |
| `unit/` | Test domain/application không cần database thật. |
| `integration/database/` | Test repository với PostgreSQL thật hoặc container test database. |
| `integration/interfaces/` | Test API qua Gin router và HTTP request. |
| `mocks/` | Mock interface cho unit test. Có thể generate bằng mockery khi số lượng interface tăng. |

### `storage/logs/`

Chỉ nên dùng cho local hoặc môi trường cần ghi file. Production ưu tiên ghi structured log ra stdout để platform thu thập.

## Request lifecycle

```text
Client
  |
  v
Gin router
  |
  v
Middleware
  |
  v
REST handler
  |  bind/validate DTO
  v
Application command/query
  |  orchestrate use case
  v
Domain entity/rule
  |
  v
Repository interface
  |
  v
Postgres repository
  |  sqlx + SQL
  v
PostgreSQL
```

Response đi ngược lại theo chiều:

```text
PostgreSQL -> repository -> application -> handler -> HTTP response -> client
```

## Bootstrap lifecycle

```text
main.go
  -> cmd/root.go
  -> cmd/http.go
  -> load config
  -> open PostgreSQL connection pool
  -> create repositories
  -> create application services/use cases
  -> create Gin handlers
  -> register middleware and routes
  -> start HTTP server
```

## Initial auth/user scope

Phase đầu chỉ cần:

```text
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET  /auth/me
PATCH /users/me
```

Các phần nên để phase sau:

- Admin user management.
- Role/permission matrix.
- Email verification.
- Forgot/reset password.
- Redis token revocation.
- Login rate limit.
- Audit log.

## Quy tắc dependency

- Handler chỉ phụ thuộc application interface.
- Application chỉ phụ thuộc domain và repository interface.
- Domain không phụ thuộc Gin, `sqlx`, PostgreSQL, Redis hoặc Kafka.
- PostgreSQL implementation chỉ nằm trong infrastructure.
- SQL không đặt trong handler/service.
- DTO HTTP không dùng trực tiếp làm domain entity.
- Không tạo package `common` hoặc `utils` để gom mọi thứ không rõ ownership.
- Chỉ thêm Kafka/Redis abstraction khi có use case thực tế.

## Implementation tasks: auth, user và Docker image

Phần này là checklist triển khai phase đầu. Estimate tính theo giờ công của một developer đã quen Go, Gin và PostgreSQL. Estimate không bao gồm triển khai production, CI/CD, email service, Redis, role/permission nâng cao hoặc audit log.

### Phase 1: Project foundation

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| FND-01 | Khởi tạo Go module và dependency | Có `go.mod`, `go.sum`, dependency cho Gin, sqlx, pgx, bcrypt, JWT và UUID; project chạy được bằng `go run` | 1.0h |
| FND-02 | Dựng package skeleton | Tạo các package `cmd`, `src/bootstrap`, `src/applications`, `src/domain`, `src/infrastructure`, `src/interfaces` theo dependency direction | 1.5h |
| FND-03 | Implement typed configuration | Đọc và validate environment variables cho server, database, JWT và runtime environment; cập nhật `.env.example` | 2.0h |
| FND-04 | Implement database lifecycle | Mở, ping, cấu hình pool và đóng `*sql.DB`/`*sqlx.DB` đúng lifecycle | 2.0h |
| FND-05 | Implement HTTP bootstrap | Wire config, database, repository, use case, handler, middleware và Gin router; server start/ graceful shutdown hoạt động | 3.0h |

**Estimate Phase 1: 9.5h**

### Phase 2: Database và domain model

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| DB-01 | Thiết kế users schema | Migration `up/down` cho bảng users, UUID, email unique, password hash, profile fields, timestamps và các constraint cần thiết | 2.0h |
| DB-02 | Thiết kế refresh token/session schema | Migration `up/down` cho refresh token đã hash, user reference, expiry, revocation và index phục vụ lookup/cleanup | 2.0h |
| DB-03 | Tạo domain entities và value objects | Có `User`, `Email` và các invariant cơ bản; domain không import Gin, sqlx hoặc PostgreSQL | 2.5h |
| DB-04 | Tạo repository interfaces | Định nghĩa interface cho user và refresh token/session theo nhu cầu application, không lộ SQL | 1.5h |

**Estimate Phase 2: 8.0h**

### Phase 3: Persistence

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| PER-01 | Implement PostgreSQL user repository | CRUD/query cần thiết: tìm theo ID/email, tạo user, cập nhật profile; xử lý `not found` và unique conflict | 3.5h |
| PER-02 | Implement refresh token repository | Tạo, tìm, revoke và xóa/đánh dấu token hết hạn; query có transaction khi cần | 3.0h |
| PER-03 | Chuẩn hóa persistence errors | Map lỗi database phổ biến thành lỗi domain/application có thể xử lý ở handler | 1.5h |
| PER-04 | Chạy và kiểm tra migrations | Apply/rollback local database, kiểm tra constraint, index, foreign key và backward compatibility | 1.5h |

**Estimate Phase 3: 9.5h**

### Phase 4: Authentication và user use cases

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| AUTH-01 | Implement password service | Hash password bằng bcrypt, verify password và không trả password hash ra response | 1.5h |
| AUTH-02 | Implement JWT service | Tạo và parse access token/refresh token, validate signature, expiry, issuer và subject theo config | 3.0h |
| AUTH-03 | Implement `Register` use case | Validate input, chuẩn hóa email, kiểm tra duplicate, hash password, tạo user và trả DTO an toàn | 2.5h |
| AUTH-04 | Implement `Login` use case | Tìm user, verify password, tạo access/refresh token và lưu refresh token/session | 2.5h |
| AUTH-05 | Implement `Refresh` use case | Validate refresh token, kiểm tra session chưa revoke/hết hạn, rotate token nếu chọn rotation và trả token pair mới | 3.0h |
| AUTH-06 | Implement `Logout` use case | Revoke refresh token hiện tại hoặc session tương ứng; thao tác idempotent | 1.5h |
| USER-01 | Implement `GetCurrentUser` use case | Lấy user từ identity trong access token và trả profile DTO, không lộ sensitive fields | 1.5h |
| USER-02 | Implement `UpdateCurrentUser` use case | Validate và cập nhật các field được phép của user hiện tại, xử lý email conflict nếu email cho phép chỉnh sửa | 2.0h |

**Estimate Phase 4: 17.5h**

### Phase 5: REST API và middleware

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| API-01 | Tạo request/response DTO | DTO riêng cho register, login, refresh, logout, user profile; không dùng DTO HTTP làm domain entity | 1.5h |
| API-02 | Implement auth handlers | Hoàn thiện `POST /auth/register`, `/login`, `/refresh`, `/logout` với status code và error response thống nhất | 3.0h |
| API-03 | Implement auth middleware | Đọc Bearer token, verify JWT, đưa authenticated user ID vào request context và xử lý unauthorized | 2.0h |
| API-04 | Implement user handlers | Hoàn thiện `GET /auth/me` và `PATCH /users/me`, bind/validate request và format response | 2.0h |
| API-05 | Chuẩn hóa middleware nền tảng | Recovery, request ID, structured logging và CORS tối thiểu cho local/API | 2.0h |
| API-06 | Đăng ký routes và health check | Route groups hoạt động, có `GET /health` hoặc tương đương và route protected được kiểm chứng | 1.0h |

**Estimate Phase 5: 11.5h**

### Phase 6: Testing

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| TST-01 | Unit test domain và security | Test email validation, password hashing/verify, JWT expiry/invalid token và các invariant chính | 2.5h |
| TST-02 | Unit test application use cases | Mock repository để test register, login, refresh, logout, current user và update user | 4.0h |
| TST-03 | Repository integration test | Chạy test với PostgreSQL test database, kiểm tra query, constraint, transaction và migration | 3.0h |
| TST-04 | HTTP integration test | Dùng `httptest` kiểm tra toàn bộ auth/user endpoints, status code, validation và protected route | 3.5h |
| TST-05 | Test commands và quality gate | `go test ./...`, `go vet ./...` và nếu có thể `golangci-lint` chạy được qua Makefile | 1.5h |

**Estimate Phase 6: 14.5h**

### Phase 7: Docker image và developer workflow

| ID | Task | Kết quả cần đạt | Estimate |
| --- | --- | --- | ---: |
| DCK-01 | Viết multi-stage Dockerfile | Build binary ở builder stage, chạy bằng runtime image nhỏ, không chứa source/dependency build thừa | 2.0h |
| DCK-02 | Thêm `.dockerignore` và runtime config | Loại trừ `.git`, local logs, test artifacts và secret; container nhận config qua environment | 0.5h |
| DCK-03 | Docker Compose local | Chạy API cùng PostgreSQL local, khai báo healthcheck, volume và network; không hard-code secret production | 2.0h |
| DCK-04 | Makefile và smoke test image | Có lệnh build/run/test/migration; build image thành công và gọi được health/auth endpoint trong container | 2.0h |
| DCK-05 | Cập nhật README | Hướng dẫn setup `.env`, migration, chạy local, test và build/run Docker image | 1.5h |

**Estimate Phase 7: 8.0h**

### Tổng estimate

| Hạng mục | Estimate |
| --- | ---: |
| Tổng effort kỹ thuật | 78.5h |
| Buffer 15% cho debugging và integration issue | 12.0h |
| Tổng dự kiến | **90.5h, khoảng 11-12 ngày làm việc** |

### Definition of done

- Các endpoint trong phần `Initial auth/user scope` hoạt động đúng với PostgreSQL.
- Password được hash bằng bcrypt; không lưu plaintext password và không trả hash trong API response.
- Access token và refresh token được validate expiry/signature; logout có thể revoke refresh session.
- Protected endpoint từ chối request thiếu hoặc có access token không hợp lệ.
- Migration `up/down` chạy được trên database local/test.
- Unit test, integration test và HTTP test chạy được bằng Makefile.
- Docker image build thành công, container start được bằng environment variables và health check trả về trạng thái healthy.
- README mô tả đầy đủ cách chạy local, migration, test và Docker.
