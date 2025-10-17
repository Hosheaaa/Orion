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
- 管理员端：HTTP Header `Authorization: Bearer <JWT>`，JWT 由后台服务基于本地密钥（HS256）签发，默认有效期 15 分钟，可通过环境变量调整。
- 演讲者/观众：后台生成的 JWT 或邀请码换取的临时令牌，由前端短期存储。
- WebSocket 握手：在 Query 或 Header 中携带 `token`、`activityId`，必要时附带 `inviteCode`。

## 3. REST API 列表

### 3.1 管理员登录
- `POST /api/v1/auth/login`
- 请求：`{ "username": "admin", "password": "***" }`
- 响应：
```json
{
  "accessToken": "...",
  "refreshToken": "...",
  "expiresIn": 900
}
```
- 说明：`expiresIn` 单位秒，对应 `ACCESS_TOKEN_TTL` 配置；刷新令牌默认有效期 7 天。

### 3.2 刷新令牌
- `POST /api/v1/auth/refresh`
- 请求：`{ "refreshToken": "..." }`
- 响应：
```json
{
  "accessToken": "...",
  "refreshToken": "...",
  "expiresIn": 900
}
```
- 说明：刷新操作会旋转刷新令牌（旧值立即失效）。

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
- 说明：默认有效期 24 小时，可在后台再次生成刷新令牌。

### 3.10 撤销单个演讲者令牌
- `POST /api/v1/activities/{id}/tokens/speaker/{tokenId}/revoke`
- 响应：`204 No Content`
- 说明：只能撤销“active” 状态的令牌，撤销后需重新生成。

### 3.11 撤销全部演讲者令牌
- `POST /api/v1/activities/{id}/tokens/speaker/revoke`
- 响应：`204 No Content`
- 说明：批量撤销该活动下所有仍有效的演讲者令牌。

### 3.12 生成观众邀请码
- `POST /api/v1/activities/{id}/tokens/viewer`
- 请求（可选字段）：`{ "maxAudience": 50, "ttlMinutes": 120 }`
- 响应：`{ "code": "ABCDE", "expiresAt": "..." }`
- 说明：未传 body 时使用默认有效期（120 分钟）且不限制观众数量；新邀请码生成后旧邀请码会被标记为 revoked。

### 3.13 查询令牌列表
- `GET /api/v1/activities/{id}/tokens`
- 响应：令牌/邀请码状态列表。

### 3.14 上传封面图片
- `POST /api/v1/uploads/cover`
- 当前状态：功能占位，接口返回 501，提示“封面上传功能尚未接入文件存储”。

### 3.15 获取语言列表
- `GET /api/v1/languages`
- 响应：`[{ "code": "en", "name": "英语" }, ...]`

### 3.16 获取观众入口二维码
- `GET /api/v1/activities/{id}/viewer-entry`
- 响应：
```json
{
  "activityId": "uuid",
  "shareUrl": "https://viewer.example.com/activity/uuid?code=ABCDE",
  "qrType": "text",
  "qrContent": "data:text/plain;base64,Li4u",
  "status": "active",
  "updatedAt": "2024-08-01T12:00:00Z"
}
```
- 说明：当前版本返回文本格式的 QR 数据 URL 占位，后续将接入二维码图片生成。

### 3.17 失效观众入口二维码
- `POST /api/v1/activities/{id}/viewer-entry/revoke`
- 响应：`{ "status": "revoked", "updatedAt": "..." }`
- 说明：活动关闭时可由系统自动调用，或管理员手动触发。

### 3.18 重新启用观众入口二维码
- `POST /api/v1/activities/{id}/viewer-entry/activate`
- 响应：`{ "status": "active", "shareUrl": "...", "qrContent": "data:text/plain;base64,...", "updatedAt": "..." }`
- 说明：若最新邀请码已过期会返回 400 并提示重新生成。
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
- 对敏感接口（生成令牌、二维码启停）建议增加速率限制与审计日志。
- JWT 采用 HS256 与本地私钥实现，需确保密钥长度与安全存储；如需多实例部署，可在后续迭代切换至非对称加签。
- 二维码链接与活动状态绑定，活动关闭后需调用 revoke 接口保证入口失效。

## 7. 版本控制
- 使用 Accept Header 或 URL 版本号（当前使用 `/api/v1`）。
- 每次对外接口变更需更新文档并提供迁移说明。
