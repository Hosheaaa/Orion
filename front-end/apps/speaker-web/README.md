# Orion Speaker Web

演讲者端控制台原型，基于 Vue3 + Vite + TypeScript + Naive UI 构建。目前聚焦以下能力：

- 活动排期选择与术语提示，展示真实描述性文字与数据指标；
- 音频采集控制台：模拟 LINEAR16 分帧推流流程，展示电平、设备与降噪状态；
- 实时字幕流、连接健康度、演讲提醒等辅助模块，满足设计规范的内容层次要求；
- Pinia 管理即时状态，TanStack Query 负责服务端数据缓存并通过 IndexedDB 持久化的准备工作（当前使用 Mock 服务，占位 API 可快捷替换）。

## 快速开始

```bash
cp .env.example .env # 配置 API / WebSocket 地址
pnpm install         # 安装依赖
pnpm --filter @orion/speaker-web dev
```

> 提示：演讲者端已直连后端 REST 与 WebSocket 接口，默认需要后台账号密码（`/api/v1/auth/login`）以及有效的 Google API Key 才能完成端到端推流。Mock 数据仅用于仪表盘辅助模块。

## 待办事项

- 接入 `@orion/shared-utils` WebSocket SDK，实现真实的音频推流与字幕回传；
- 绑定后台活动接口 `/api/v1/activities`、字幕历史 `/api/v1/subtitles`；
- 将“通知技术值班”等操作连接后台事件上报 API；
- 添加端到端测试（Cypress/Playwright）覆盖推流、字幕与连接异常场景。
