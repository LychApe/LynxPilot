# LynxPilot

使用 Gin + GORM，并默认使用 SQLite（数据库文件默认 `config/lynxpilot.db`，可通过环境变量 `SQLITE_PATH` 覆盖）。
应用配置在 `config/config.yaml`，包括服务端口与 token 盐（可通过 `CONFIG_PATH` 覆盖配置文件路径）。

启动服务：

```bash
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```
