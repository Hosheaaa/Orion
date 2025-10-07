# API 规范

## 1. 总则
- 所有 REST API 前缀为 `/api/v1`。
- 请求与响应统一使用 JSON，编码 UTF-8。
- 时间字段采用 ISO 8601 格式，使用 UTC 时间。
- 错误响应格式：
```json
{
  "code": "ACTIVITY_NOT_FOUND",
  "message": "活动不存在",
  "data": null
}
```

## 2. 鉴权机制
- 管理员端：HTTP Header `Authorization: Bearer <JWT>`。
- 演讲者/观众：后台生成的 JWT 或邀请码换取的临时令牌，由前端短期存储。
- WebSocket 握手：在 Query 或 Header 中携带 `token`、`activityId`，必要时附带 `inviteCode`。

## 3. REST API 列表

### 3.1 管理员登录
- `POST /api/v1/auth/login`
- 请求：`{ "username": "admin", "password": "***" }`
- 响应：`{ "accessToken": "...", "refreshToken": "...", "expiresIn": 7200 }`

### 3.2 刷新令牌
- `POST /api/v1/auth/refresh`
- 请求：`{ "refreshToken": "..." }`
- 响应：`{ "accessToken": "...", "expiresIn": 7200 }`

### 3.3 活动列表
- `GET /api/v1/activities`
- 查询参数：`status`（可选：draft/published/closed）
- 响应：`[{ "id": "uuid", "title": "...", "status": "published", ... }]`

### 3.4 创建活动
- `POST /api/v1/activities`
- 请求：
```json
{
  "title": "新品发布会",
  "description": "介绍产品亮点",
  "speaker": "张三",
  "startTime": "2024-05-01T12:00:00Z",
  "inputLanguage": "zh-CN",
  "targetLanguages": ["en", "ja", "es"],
  "coverUrl": "https://.../cover.png"
}
```
- 响应：活动详情（含观众端链接与二维码 Base64 数据）。

### 3.5 更新活动
- `PUT /api/v1/activities/{id}`
- 请求体同创建。
- 响应：更新后的活动详情。

### 3.6 发布/关闭活动
- `POST /api/v1/activities/{id}/publish`
- `POST /api/v1/activities/{id}/close`
- 响应：`{ "id": "uuid", "status": "published" }`
- 说明：活动关闭时，后台自动标记观众二维码失效。

### 3.7 获取单个活动详情
- `GET /api/v1/activities/{id}`
- 响应：活动详细信息，包含语种配置、令牌信息摘要、观众入口信息（链接、二维码状态）。

### 3.8 删除活动
- `DELETE /api/v1/activities/{id}`（软删除或受限操作，需确认业务政策）。

### 3.9 生成演讲者令牌
- `POST /api/v1/activities/{id}/tokens/speaker`
- 响应：`{ "token": "jwt", "expiresAt": "..." }`

### 3.10 生成观众邀请码
- `POST /api/v1/activities/{id}/tokens/viewer`
- 请求：`{ "maxAudience": 50, "ttlMinutes": 120 }`
- 响应：`{ "code": "ABCDE", "expiresAt": "..." }`

### 3.11 查询令牌列表
- `GET /api/v1/activities/{id}/tokens`
- 响应：令牌/邀请码状态列表。

### 3.12 上传封面图片
- `POST /api/v1/uploads/cover`
- 请求：`multipart/form-data`，字段 `file`
- 响应：`{ "url": "https://..." }`

### 3.13 获取语言列表
- `GET /api/v1/languages`
- 响应：`[{ "code": "en", "name": "英语" }, ...]`

### 3.14 获取观众入口二维码
- `GET /api/v1/activities/{id}/viewer-entry`
- 查询参数：`format`（可选：`svg`/`png`/`base64`，默认 `png`）。
- 响应：
```json
{
  "shareUrl": "https://viewer.example.com/activity/uuid?code=ABCDE",
  "qrType": "png",
  "qrContent": "data:image/png;base64,....",
  "status": "active"
}
```

### 3.15 失效观众入口二维码
- `POST /api/v1/activities/{id}/viewer-entry/revoke`
- 请求：`{ "reason": "MANUAL" }`
- 响应：`{ "status": "revoked" }`
- 说明：活动关闭时可由系统自动调用，或管理员手动触发。

### 3.16 重新启用观众入口二维码
- `POST /api/v1/activities/{id}/viewer-entry/activate`
- 响应：`{ "status": "active", "qrContent": "data:image/png;base64,..." }`
- 说明：仅在活动重新开放时使用。

## 4. WebSocket 接口

### 4.1 演讲者通道
- URL：`wss://domain/ws/speaker`
- 参数：`token`, `activityId`
- 消息示例：
```json
{"type":"AUTH","payload":{"activityId":"uuid","lang":"zh-CN"}}
```
- 音频数据：二进制帧（推荐使用 protobuf 或 JSON+Base64 定义格式），示例：
```json
{"type":"AUDIO","payload":{"chunk":"BASE64", "sequence":123}}
```
- 控制消息：
```json
{"type":"CONTROL","payload":{"action":"STOP"}}
```
- 服务端响应：
```json
{"type":"STATE","payload":{"status":"READY"}}
```

### 4.2 观众通道
- URL：`wss://domain/ws/viewer`
- 参数：`token`, `activityId`, `lang`
- 观众扫码进入后，通过邀请码换取 `token`，再建立连接。
- 服务端发送 `SUBTITLE`：
```json
{
  "type": "SUBTITLE",
  "payload": {
    "sentenceId": "uuid",
    "original": "大家好",
    "lang": "en",
    "text": "Hello everyone",
    "timestamp": "2024-05-01T12:00:03Z"
  }
}
```
- 历史字幕：连接成功后发送 `HISTORY` 包含最近 5 分钟的数组。
- 错误状态：
```json
{"type":"STATE","payload":{"status":"ERROR","message":"活动已结束"}}
```

## 5. 错误码
| 错误码 | 含义 | HTTP 状态 |
| --- | --- | --- |
| `UNAUTHORIZED` | 认证失败 | 401 |
| `FORBIDDEN` | 无权限访问资源 | 403 |
| `ACTIVITY_NOT_FOUND` | 活动不存在 | 404 |
| `ACTIVITY_CLOSED` | 活动已关闭 | 409 |
| `INVALID_LANGUAGE` | 目标语言不受支持 | 400 |
| `GOOGLE_STT_ERROR` | Google STT 调用失败 | 502 |
| `GOOGLE_TRANSLATE_ERROR` | 翻译调用失败 | 502 |
| `QR_GENERATE_FAILED` | 二维码生成失败 | 500 |
| `RATE_LIMITED` | 接口调用频率过高 | 429 |

## 6. 安全要求
- 所有接口必须通过 HTTPS/WSS。
- 对敏感接口（生成令牌、上传资源、二维码失效/启用）增加速率限制与审计日志。
- JWT 使用非对称密钥签名（RS256），便于多实例校验。
- 二维码链接与活动状态绑定，活动关闭后自动失效并返回提示页面。

## 7. 版本控制
- 使用 Accept Header 或 URL 版本号（当前使用 `/api/v1`）。
- 每次对外接口变更需更新文档并提供迁移说明。
