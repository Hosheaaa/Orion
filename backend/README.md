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
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
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

## 开发指南

### 添加新的 API 端点

1. 在 `internal/api/handler/` 中创建处理器函数
2. 在 `internal/api/router.go` 中注册路由
3. 如需业务逻辑，在 `internal/app/` 中实现服务
4. 如需数据模型，在 `internal/domain/` 中定义实体

### 环境变量说明

- `APP_PORT`: 服务器端口（默认 8080）
- `APP_ENV`: 运行环境（development/production）
- `JWT_SECRET_PATH`: JWT 私钥文件路径
- `GOOGLE_APPLICATION_CREDENTIALS`: Google 服务账户凭证文件路径
- `REDIS_URL`: Redis 连接 URL
- `VIEWER_BASE_URL`: 观众端基础 URL（用于生成二维码）

## 下一步

- [ ] 实现 JWT 认证逻辑
- [ ] 实现活动管理服务
- [ ] 集成 Google Speech-to-Text API
- [ ] 集成 Google Translation API
- [ ] 实现 WebSocket 处理
- [ ] 实现二维码生成
