# LynxPilot

基于 Go 的后端 Web 服务，提供 RESTful API 和进程自管理能力。

## 技术栈

- Go 1.26
- [Gin](https://github.com/gin-gonic/gin) — HTTP 框架
- [GORM](https://gorm.io) + SQLite — ORM 与存储
- YAML — 配置文件

## 项目结构

```
├── cmd/server/main.go          # 程序入口
├── internal/
│   ├── bootstrap/              # 启动引导（配置、数据库、路由初始化）
│   ├── api/                    # HTTP Handler（按领域分包）
│   │   ├── server/             # 服务控制 API（reboot/shutdown）
│   │   └── user/               # 用户 API（login）
│   ├── router/                 # 路由注册（按领域分包）
│   │   ├── server/
│   │   └── user/
│   ├── service/process/        # 进程管理（自重启逻辑）
│   └── utils/logger/           # 日志工具
└── config/                     # 运行时配置（config.yaml、SQLite 数据库）
```

分层架构：`router`（路由注册）→ `api`（Handler）→ `service`（业务逻辑）。

## API

当前路由前缀统一为 `/api/v1/public/`。

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/public/server/reboot` | 重启服务（拉起新进程后退出当前进程） |
| GET | `/api/v1/public/server/shutdown` | 关闭服务 |
| POST | `/api/v1/public/user/login` | 用户登录（占位） |

## 配置

配置文件路径：`config/config.yaml`（已被 gitignore，需自行创建）。

```yaml
server:
  port: 8080          # 1-65535
  mode: release       # release | debug

auth:
  token_salt: "your-secret-key"

database:
  path: config/lynxpilot.db
```

配置加载支持多级目录回退搜索，并使用严格模式校验未知字段。

## 运行

```bash
# 开发模式
go run ./cmd/server/main.go

# 编译后运行
go build -o lynxpilot ./cmd/server/main.go
./lynxpilot
```

服务启动后监听配置中指定的端口，支持通过 API 触发重启和优雅关闭（SIGINT/SIGTERM，5s 超时）。

## 进程自重启

`/api/v1/public/server/reboot` 的实现逻辑：

1. 返回 200 响应给客户端
2. 异步拉起新进程（继承当前进程的 stdout/stderr/env）
3. 向当前进程发送 `os.Interrupt` 信号退出

在 `go run` 开发模式下，自动检测临时编译产物并回退为 `go run ./cmd/server/main.go`。
