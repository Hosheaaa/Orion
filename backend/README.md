# Orion Backend

Orion 实时翻译系统后端服务。

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入必要的配置
```

### 运行服务

```bash
# 开发模式
go run cmd/server/main.go

# 编译并运行
go build -o bin/orion-server cmd/server/main.go
./bin/orion-server
```

### 测试

```bash
# 运行所有测试
GOCACHE=$(pwd)/.gocache go test ./...

# 运行测试并显示覆盖率
GOCACHE=$(pwd)/.gocache go test -cover ./...

# 生成覆盖率报告
GOCACHE=$(pwd)/.gocache go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 项目结构

```
backend/
├── cmd/
│   └── server/          # 应用入口
│       └── main.go
├── internal/
│   ├── api/             # API 层
│   │   ├── handler/     # 请求处理器
│   │   ├── middleware/  # 中间件
│   │   └── router.go    # 路由配置
│   ├── app/             # 应用服务层
│   ├── domain/          # 领域层
│   └── infra/           # 基础设施层
│       └── config/      # 配置管理
├── pkg/                 # 可复用的公共包
├── configs/             # 配置文件
├── go.mod
└── README.md
```

## API 文档

启动服务后访问健康检查接口：

```bash
curl http://localhost:8080/health
```

完整 API 文档请参考 `/docs/API 规范.md`

### WebSocket 鉴权与接入

- 演讲者端：`/ws/speaker?activityId=<id>&token=<speaker-token>&language=<sourceLang>`，`token` 由管理端接口生成，仅限草稿/已发布活动，语言必须与活动输入语种一致。
- 观众端：`/ws/viewer?activityId=<id>&token=<viewer-code>&language=<targetLang>`，活动需处于已发布状态，`language` 必须在活动目标语言列表内。
- 所有令牌由 `AccessService` 统一校验并自动失效过期或撤销的令牌；如需扩展鉴权，可替换仓储实现以接入数据库。

### CORS 配置说明

- `CORS_ALLOWED_ORIGINS` 支持逗号分隔的白名单列表，例如：`https://console.orion.com,https://admin.orion.com`。
- 若仅用于本地调试，可保持默认值 `http://localhost:3000`；生产环境务必显式列出可信前端域名。
- 服务启动后会基于白名单回显 `Access-Control-Allow-Origin`，同时允许携带 Cookie、Authorization 等凭证，避免使用 `*` 带来的安全风险。

## 开发指南

### 添加新的 API 端点

1. 在 `internal/api/handler/` 中创建处理器函数
2. 在 `internal/api/router.go` 中注册路由
3. 如需业务逻辑，在 `internal/app/` 中实现服务
4. 如需数据模型，在 `internal/domain/` 中定义实体

### 环境变量说明

- `APP_PORT`: 服务器端口（默认 8080）
- `APP_ENV`: 运行环境（development/production）
- `ADMIN_USERNAME`: 管理后台账号（默认 admin）
- `ADMIN_PASSWORD`: 管理后台密码（默认 admin123，建议运行前立即修改）
- `ACCESS_TOKEN_TTL`: 访问令牌有效期（默认 15m）
- `REFRESH_TOKEN_TTL`: 刷新令牌有效期（默认 168h）
- `JWT_SECRET_PATH`: JWT 私钥文件路径
- `JWT_SECRET_PATH` 指向的文件需包含至少 32 个字符的随机字符串，用于 HMAC 签名
- `CORS_ALLOWED_ORIGINS`: CORS 白名单地址，多个域名使用逗号分隔
- `GOOGLE_APPLICATION_CREDENTIALS`: Google 服务账户凭证文件路径
- `REDIS_URL`: Redis 连接 URL
- `VIEWER_BASE_URL`: 观众端基础 URL（用于生成二维码）
- `GOOGLE_STT_API_KEY` / `GOOGLE_TRANSLATE_API_KEY`: 启用实时翻译所需的 Google API Key，缺失时 WebSocket 功能将返回 503。

## 下一步

- [ ] 将活动、令牌持久化至数据库或缓存
- [ ] 基于 AccessService 接入更严格的令牌生命周期审计
- [ ] 集成 Google Speech-to-Text API
- [ ] 集成 Google Translation API
- [ ] 实现二维码生成
